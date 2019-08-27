import numpy as np
import PIL as pillow
from PIL import Image, ImageDraw, ImageEnhance

def hilbert_curve_filter(image):
  f = ImageEnhance.Brightness(pillow.Image.fromarray(image.astype(np.uint8), "RGB"))
  f = ImageEnhance.Contrast(f.enhance(1.3)).enhance(10)
  im = np.array(f)
  k = 1
  while k < max(image.shape[0], image.shape[1]):
    k *= 2
  W = k - 1
  himage = pillow.Image.new("RGB", (W * 2, W * 2), (255, 255, 255))
  draw = pillow.ImageDraw.Draw(himage)
  points = hilbert_curve(*map(np.array, [[0, 0], [W, 0], [W, W], [0, W]]), np.array(f))
  points = list(filter(lambda x: x is not None, points))
  draw.line(list(map(lambda p: (int(p[1]) * 2, int(p[0]) * 2), points)), fill=(0, 0, 0))
  return himage.crop((0, 0, image.shape[1] * 2, image.shape[0] * 2))

def is_block_black(p, q, im):
  THRESHOLD = 0.6
  f = lambda x: int(np.floor(x))
  l, r = map(f, sorted([p[0], q[0]]))
  t, b = map(f, sorted([p[1], q[1]]))
  c = 0
  for i in range(l, r):
    for j in range(t, b):
      c += np.sum(im[min(i, im.shape[0] - 1)][min(im.shape[1] - 1, j)]) / (255 * 3)
  return c < THRESHOLD * (b - t) * (r - l)

def hilbert_curve(p1, p2, p3, p4, im):
  n = int(np.sqrt((p2 - p1).dot(p2 - p1))) // 2
  if n == 0:
    if is_block_black(p1, p3, im):
      return [p1, p2, p3, p4]
    else:
      return [None]
  p12 = (p2 - p1) / (2 * n + 1)
  p23 = (p3 - p2) / (2 * n + 1)
  # 1       4
  # |1-2 3-4|
  # | a| |d |
  # |4-3 2-1|
  # ||  p  ||
  # |1 4-1 4|
  # ||b| |c||
  # |2-3 2-3|
  # 2-------3
  a1, a2, a3, a4 = p1, p1 + p23 * n, p1 + p23 * n + p12 * n, p1 + p12 * n
  b1, b2, b3, b4 = p2 - p12 * n, p2, p2 + p23 * n, p2 + p23 * n - p12 * n
  c1, c2, c3, c4 = p3 - p23 * n - p12 * n, p3 - p23 * n, p3, p3 - p12 * n
  d1, d2, d3, d4 = p4 + p12 * n, p4 + p12 * n - p23 * n, p4 - p23 * n, p4
  resa = hilbert_curve(a1, a2, a3, a4, im)
  resb = hilbert_curve(b1, b2, b3, b4, im)
  resc = hilbert_curve(c1, c2, c3, c4, im)
  resd = hilbert_curve(d1, d2, d3, d4, im)
  return resa + resb + resc + resd