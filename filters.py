from time import asctime

import requests
import numpy as np
from PIL import Image
from sklearn.cluster import KMeans

from convolution import apply_filter
from hilbert import hilbert_curve_filter

def get_next_filename():
    return str(asctime())

def load_image(url):
    imid = get_next_filename()
    image_filename = "img/" + imid + ".orig.png"
    with open(image_filename, "wb") as f:
        f.write(requests.get(url).content)
    im = Image.open(image_filename)
    return np.array(im), imid

def apply_convolution(url, kernel):
    image, imid = load_image(url)
    filtered = apply_filter(image, kernel)
    filtered_filename = "img/" + imid + ".res.png"
    filtered.save(filtered_filename)
    return filtered_filename

def cluster_filter(url, N):
    image, imid = load_image(url)
    X = []
    for i in range(image.shape[0]):
        for j in range(image.shape[1]):
            X.append(image[i][j])
    X = np.array(X)
    kmeans = KMeans(n_clusters=N, random_state=0).fit(X)
    print("Clusters")
    colors = kmeans.cluster_centers_
    r = np.zeros(image.shape)
    for i in range(r.shape[0]):
        for j in range(r.shape[1]):
            d = [np.dot(image[i][j] - color, image[i][j] - color) for color in colors]
            k = 0
            for h in range(1, len(colors)):
                if d[h] < d[k]:
                    k = h
            r[i][j] = colors[k]
    filtered = Image.fromarray(r.astype(np.uint8), "RGB")
    filtered_filename = "img/" + imid + ".res.png"
    filtered.save(filtered_filename)
    return filtered_filename

def hilbert_curve(url):
    image, imid = load_image(url)
    tmp = hilbert_curve_filter(image)
    filtered_filename = "img/" + imid + ".res.png"
    tmp.save(filtered_filename)
    return filtered_filename
