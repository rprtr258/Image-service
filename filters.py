from time import asctime

import requests
import numpy as np
from PIL import Image
from sklearn.cluster import KMeans

from convolution import apply_filter
from hilbert import hilbert_curve_filter

def get_next_filename():
    return "img/" + str(asctime()) + ".png"

def load_image(url):
    with open("img/f.jpg", "wb") as f:
        f.write(requests.get(url).content)
    im = Image.open("img/f.jpg")
    return np.array(im)

def apply_convolution(url, kernel):
    image = load_image(url)
    filtered = apply_filter(image, kernel)
    filtered_filename = get_next_filename()
    filtered.save(filtered_filename)
    return filtered_filename

def cluster_filter(url, N):
    image = load_image(url)
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
    filtered_filename = get_next_filename()
    filtered.save(filtered_filename)
    return filtered_filename

def hilbert_curve(url):
    image = load_image(url)
    tmp = hilbert_curve_filter(image)
    res_filename = get_next_filename()
    tmp.save(res_filename)
    return res_filename
