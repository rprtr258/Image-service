#!/usr/bin/env python3
from app import app
from routes import *

if __name__ == "__main__":
    app.run("0.0.0.0", 80, debug=True)
