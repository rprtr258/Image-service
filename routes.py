from flask import render_template

from app import app

@app.errorhandler(404)
def not_found(error):
    return render_template("404.html"), 404

@app.route("/", methods=["GET"])
def index():
    return render_template("index.html")

@app.route("/blur", methods=["GET", "POST"])
def blur():
    return render_template("blur.html")
