from http.client import HTTPException
import inspect
import json
import mimetypes
import pathlib
import sys
import tempfile
import time
import types
from urllib.parse import urlparse
from fastapi import FastAPI, Response
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from typing import Literal, Type

from flask import Flask
from itsdangerous import base64_decode
from pydantic import BaseModel, Field

from ..input import (
    validate_and_convert_inputs,
    InputValidationError,
)
import cog
from ..json import to_json
from ..predictor import Predictor, run_prediction, load_predictor


class JSONEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, cog.File):
            return cog.File.encode(obj)
        elif isinstance(obj, cog.Path):
            return cog.Path.encode(obj)
        return json.JSONEncoder.default(self, obj)


def create_app(predictor: Predictor) -> FastAPI:
    app = FastAPI()
    app.on_event("startup")(predictor.setup)

    predict_types = inspect.getfullargspec(predictor.predict).annotations
    InputType = predict_types.get("input")
    OutputType = predict_types.get("return", Literal[None])

    class ResponseData(BaseModel):
        status: str = Field(...)
        output: OutputType = Field(...)

        class Config:
            # json_dumps = lambda *args, **kwargs: json.dumps(
            #     *args, cls=JSONEncoder, **kwargs
            # )
            # json_dumps = lambda _: "foo"
            json_encoders = {cog.Path: lambda _: "foo"}

        # json_dumps = lambda _: "foo"

    def create_response(output) -> ResponseData:
        res = ResponseData(status="success", output=output)

        # HACK: https://github.com/tiangolo/fastapi/pull/2061
        return Response(
            content=json.dumps(res.dict(), cls=JSONEncoder),
            media_type="application/json",
        )

    if InputType:

        def predict(input: InputType):
            return create_response(predictor.predict(input))

    else:

        def predict():
            return create_response(predictor.predict())

    app.post("/predict", response_model=ResponseData)(predict)

    return app


class HTTPServer:
    def __init__(self, predictor: Predictor):
        self.predictor = predictor

    def make_app(self) -> Flask:
        start_time = time.time()
        self.predictor.setup()
        app = Flask(__name__)
        setup_time = time.time() - start_time

        @app.route("/predict", methods=["POST"])
        @app.route("/infer", methods=["POST"])  # deprecated
        def handle_request():
            start_time = time.time()

            cleanup_functions = []
            try:
                raw_inputs = {}
                for key, val in request.form.items():
                    raw_inputs[key] = val
                for key, val in request.files.items():
                    if key in raw_inputs:
                        return _abort400(
                            f"Duplicated argument name in form and files: {key}"
                        )
                    raw_inputs[key] = val

                if hasattr(self.predictor.predict, "_inputs"):
                    try:
                        inputs = validate_and_convert_inputs(
                            self.predictor, raw_inputs, cleanup_functions
                        )
                    except InputValidationError as e:
                        return _abort400(str(e))
                else:
                    inputs = raw_inputs

                result = run_prediction(self.predictor, inputs, cleanup_functions)
                run_time = time.time() - start_time
                return self.create_response(result, setup_time, run_time)
            finally:
                for cleanup_function in cleanup_functions:
                    try:
                        cleanup_function()
                    except Exception as e:
                        sys.stderr.write(f"Cleanup function caught error: {e}")

        @app.route("/ping")
        def ping():
            return "PONG"

        @app.route("/type-signature")
        def type_signature():
            return jsonify(self.predictor.get_type_signature())

        return app

    def start_server(self):
        app = self.make_app()
        app.run(host="0.0.0.0", port=5000, threaded=False, processes=1)

    def create_response(self, result, setup_time, run_time):
        # loop over generator function to get the last result
        if isinstance(result, types.GeneratorType):
            last_result = None
            for iteration in enumerate(result):
                last_result = iteration
            # last result is a tuple with (index, value)
            result = last_result[1]

        if isinstance(result, Path):
            resp = send_file(str(result))
        elif isinstance(result, str):
            resp = Response(result, mimetype="text/plain")
        else:
            resp = Response(to_json(result), mimetype="application/json")
        resp.headers["X-Setup-Time"] = setup_time
        resp.headers["X-Run-Time"] = run_time
        return resp


def _abort400(message):
    resp = jsonify({"message": message})
    resp.status_code = 400
    return resp


if __name__ == "__main__":
    predictor = load_predictor()
    server = HTTPServer(predictor)
    server.start_server()
