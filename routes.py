from flask import render_template, request
import requests

from app import app

@app.errorhandler(404)
def not_found(error):
    return render_template("404.html"), 404

@app.route("/", methods=["GET"])
def index():
    return render_template("index.html")

@app.route("/blur", methods=["GET", "POST"])
def blur():
    if request.method == "POST":
        if "url" in request.form:
            try:
                resp = requests.get(request.form["url"])
                if resp.status_code != 200:
                    return render_template("blur.html", message="Incorrect url")
                return render_template("blur.html", message=f"Processed image {request.form['url']}")
            except Exception:
                return render_template("blur.html", message="Incorrect url")
        else:
            return render_template("blur.html", message="Url was not provided")
    else:
        return render_template("blur.html")
