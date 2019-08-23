import requests
import numpy as np
from PIL import Image

from convolution import apply_filter

def load_image(url):
    with open("f.jpg", "wb") as f:
        f.write(requests.get(url).content)
    im = Image.open("f.jpg")
    return np.array(im)

def apply_convolution(url, kernel):
    image = load_image(url)
    filtered = apply_filter(image, kernel)
    filtered_filename = "img/g.png"
    filtered.save(filtered_filename)
    return filtered_filename
