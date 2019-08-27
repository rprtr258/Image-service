from flask import render_template, request, send_from_directory

from filters import apply_convolution, cluster_filter, hilbert_curve, hilbert_darken
from app import app

def convolution_filter(kernel, filter_name):
    if request.method == "POST":
        if "url" in request.form:
            try:
                image_file = apply_convolution(request.form["url"], kernel)
                return render_template("filter.html", message=f"Processed image {request.form['url']}", image_file=image_file, filter_name=filter_name)
            except Exception as e:
                return render_template("filter.html", message=f"Error occured:\n{e}", filter_name=filter_name)
        else:
            return render_template("filter.html", message="Url was not provided", filter_name=filter_name)
    else:
        return render_template("filter.html", filter_name=filter_name)

@app.errorhandler(404)
def not_found(error):
    return render_template("404.html"), 404

@app.route('/img/<path:path>')
def send_img(path):
    return send_from_directory('img', path)

@app.route("/", methods=["GET"])
def index():
    return render_template("index.html")

@app.route("/blur", methods=["GET", "POST"])
def blur():
    return convolution_filter([[1, 1, 1], [1, 1, 1], [1, 1, 1]], "Blur")

@app.route("/weakblur", methods=["GET", "POST"])
def weakblur():
    return convolution_filter([[0, 1, 0], [1, 1, 1], [0, 1, 0]], "Weak blur")

@app.route("/emboss", methods=["GET", "POST"])
def emboss():
    return convolution_filter([[-2, -1, 0], [-1, 1, 1], [0, 1, 2]], "Emboss")

@app.route("/sharpen", methods=["GET", "POST"])
def sharpen():
    return convolution_filter([[0, -1, 0], [-1, 5, -1], [0, -1, 0]], "Sharpen")

@app.route("/edgeenhance", methods=["GET", "POST"])
def edgeenhance():
    return convolution_filter([[0, 0, 0], [-1, 1, 0], [0, 0, 0]], "Edge enhance")

@app.route("/edgedetect1", methods=["GET", "POST"])
def edgedetect1():
    return convolution_filter([[1, 0, -1], [0, 0, 0], [-1, 0, 1]], "Edge detect 1")

@app.route("/edgedetect2", methods=["GET", "POST"])
def edgedetect2():
    return convolution_filter([[0, -1, 0], [-1, 4, -1], [0, -1, 0]], "Edge detect 2")

@app.route("/horizontallines", methods=["GET", "POST"])
def horizontallines():
    return convolution_filter([[-1, -1, -1], [2, 2, 2], [-1, -1, -1]], "Horizontal lines")

@app.route("/verticallines", methods=["GET", "POST"])
def verticallines():
    return convolution_filter([[-1, 2, -1], [-1, 2, -1], [-1, 2, -1]], "Vertical lines")

@app.route("/cluster", methods=["GET", "POST"])
def cluster():
    if request.method == "POST":
        if "url" in request.form:
            try:
                image_file = cluster_filter(request.form["url"], int(request.form["n"]))
                return render_template("cluster.html", message=f"Processed image {request.form['url']}", image_file=image_file)
            except Exception as e:
                return render_template("cluster.html", message=f"Error occured:\n{e}")
        else:
            return render_template("cluster.html", message="Url was not provided")
    else:
        return render_template("cluster.html")

@app.route("/hilbert", methods=["GET", "POST"])
def hilbert():
    if request.method == "POST":
        if "url" in request.form:
            try:
                image_file = hilbert_curve(request.form["url"])
                return render_template("filter.html", message=f"Processed image {request.form['url']}", image_file=image_file)
            except Exception as e:
                return render_template("filter.html", message=f"Error occured:\n{e}")
        else:
            return render_template("filter.html", message="Url was not provided")
    else:
        return render_template("filter.html")

@app.route("/hilbertdarken", methods=["GET", "POST"])
def hilbertdarken():
    if request.method == "POST":
        if "url" in request.form:
            try:
                image_file = hilbert_darken(request.form["url"])
                return render_template("filter.html", message=f"Processed image {request.form['url']}", image_file=image_file)
            except Exception as e:
                return render_template("filter.html", message=f"Error occured:\n{e}")
        else:
            return render_template("filter.html", message="Url was not provided")
    else:
        return render_template("filter.html")
