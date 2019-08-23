import numpy as np
import PIL as pillow

__all__ = ["apply_filter"]

def get_projections(f):
  f_height, f_width = f.shape[0], f.shape[1]
  fr, fg, fb = np.zeros((f_height, f_width)), np.zeros((f_height, f_width)), np.zeros((f_height, f_width))
  for i in range(f_height):
    for j in range(f_width):
      fr[i][j], fg[i][j], fb[i][j] = f[i][j][0], f[i][j][1], f[i][j][2]
  return fr, fg, fb

def convolution(f, g):
  f_height, f_width = f.shape[0], f.shape[1]
  g_height, g_width = g.shape[0], g.shape[1]
  fc = np.zeros((g_height - 1 + f_height + g_height - 1, g_width - 1 + f_width + g_width - 1))
  for i in range(f_height):
    for j in range(f_width):
      fc[i + g_height - 1][j + g_width - 1] = f[i][j]
  r = np.zeros((g_height - 1 + f_height + g_height - 1, g_width - 1 + f_width + g_width - 1))
  for i in range(g_height - 1 + f_height):
    for j in range(g_width - 1 + f_width):
      r[i][j] += np.multiply(fc[np.ix_(range(i, i + g_height), range(j, j + g_width))], g).sum()
  return r

def convolution3D(f, g):
  fr, fg, fb = get_projections(f)
  gr, gg, gb = get_projections(g)
  rr, rg, rb = convolution(fr, gr), convolution(fg, gg), convolution(fb, gb)
  r = np.zeros((rr.shape[0], rr.shape[1], 3))
  for i in range(r.shape[0]):
    for j in range(r.shape[1]):
      r[i][j][0] = rr[i][j]
      r[i][j][1] = rg[i][j]
      r[i][j][2] = rb[i][j]
  return r

def thrice(f):
  r = np.zeros((f.shape[0], f.shape[1], 3))
  for i in range(f.shape[0]):
    for j in range(f.shape[1]):
      r[i][j][0] = f[i][j]
      r[i][j][1] = f[i][j]
      r[i][j][2] = f[i][j]
  return r

def apply_npfilter(f, m):
  r = convolution3D(f, thrice(m))
  r = (r - np.ones(r.shape) * np.min(r)) / (np.max(r) - np.min(r)) * 255
  return pillow.Image.fromarray(r.astype(np.uint8), "RGB")

def apply_filter(f, m):
  return apply_npfilter(f, np.array(m))
