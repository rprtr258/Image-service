#!/usr/bin/env python3
from app import app
from routes import *

# TODO: cache files by url
# TODO: save jpgs as jpg, not png
# TODO: offload work to workers
if __name__ == "__main__":
    app.run("0.0.0.0", 80, debug=True)
