import base64
from enum import Enum
import io
import os
from pathlib import Path
import tempfile

from PIL import Image
from pydantic import BaseModel, Field
import pytest

import cog
from .test_http import make_client


def test_no_input():
    class Predictor(cog.Predictor):
        def predict(self) -> str:
            return "foobar"

    client = make_client(Predictor())
    resp = client.post("/predict")
    assert resp.status_code == 200
    assert resp.json() == {"status": "success", "output": "foobar"}


def test_good_str_input():
    class Input(BaseModel):
        text: str

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> str:
            return input.text

    client = make_client(Predictor())
    resp = client.post("/predict", json={"text": "baz"})
    assert resp.status_code == 200
    assert resp.json() == {"status": "success", "output": "baz"}


def test_good_int_input():
    class Input(BaseModel):
        num: int

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> int:
            return input.num ** 3

    client = make_client(Predictor())
    resp = client.post("/predict", json={"num": 3})
    assert resp.status_code == 200
    assert resp.json() == {"output": 27, "status": "success"}
    resp = client.post("/predict", json={"num": -3})
    assert resp.status_code == 200
    assert resp.json() == {"output": -27, "status": "success"}


def test_bad_int_input():
    class Input(BaseModel):
        num: int

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> int:
            return input.num ** 2

    client = make_client(Predictor())
    resp = client.post("/predict", json={"num": "foo"})
    assert resp.status_code == 422


def test_default_int_input():
    class Input(BaseModel):
        num: int = Field(5)

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> int:
            return input.num ** 2

    client = make_client(Predictor())

    resp = client.post("/predict", json={})
    assert resp.status_code == 200
    assert resp.json() == {"output": 25, "status": "success"}

    resp = client.post("/predict", json={"num": 3})
    assert resp.status_code == 200
    assert resp.json() == {"output": 9, "status": "success"}


def test_file_input_data_url():
    class Predictor(cog.Predictor):
        def predict(self, file: cog.File) -> Path:
            return input.file.read()

    client = make_client(Predictor())
    resp = client.post(
        "/predict",
        json={
            "file": "data:text/plain;base64," + base64.b64encode(b"bar").decode("utf-8")
        },
    )
    assert resp.json() == {"output": "bar", "status": "success"}
    assert resp.status_code == 200


def test_path_input_data_url():
    class Input(BaseModel):
        path: cog.Path

    class Predictor(cog.Predictor):
        def setup(self):
            pass

        def predict(self, input: Input) -> str:
            with open(input.path) as fh:
                extension = fh.name.split(".")[-1]
                return f"{extension} {fh.read()}"

    client = make_client(Predictor())
    resp = client.post(
        "/predict",
        json={
            "path": "data:text/plain;base64," + base64.b64encode(b"bar").decode("utf-8")
        },
    )
    assert resp.json() == {"output": "txt bar", "status": "success"}
    assert resp.status_code == 200


def test_file_bad_input():
    class Input(BaseModel):
        file: cog.File

    class Predictor(cog.Predictor):
        def setup(self):
            pass

        def predict(self, input: Input) -> str:
            return input.file.read()

    client = make_client(Predictor())
    resp = client.post(
        "/predict",
        json={"file": "foo"},
    )
    assert resp.status_code == 422


def test_path_output_file():
    class Predictor(cog.Predictor):
        def setup(self):
            pass

        def predict(self) -> cog.Path:
            temp_dir = tempfile.mkdtemp()
            temp_path = os.path.join(temp_dir, "my_file.bmp")
            img = Image.new("RGB", (255, 255), "red")
            img.save(temp_path)
            return cog.Path(temp_path)

    client = make_client(Predictor())
    res = client.post("/predict")
    assert res.status_code == 200
    header, b64data = res.json()["output"].split(",", 1)
    assert header == "data:image/bmp;base64"
    assert len(base64.b64decode(b64data)) == 195894


def test_extranous_input_keys():
    class Input(BaseModel):
        text: str

    class Predictor(cog.Predictor):
        def predict(self, input: Input):
            return input.text

    client = make_client(Predictor())
    resp = client.post("/predict", json={"text": "baz", "text2": "qux"})
    assert resp.status_code == 422


def test_multiple_arguments():
    class Input(BaseModel):
        text: str
        num1: int
        num2: int = Field(10)
        path: cog.Path

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> str:
            with open(input.path) as fh:
                return input.text + " " + str(input.num1 * input.num2) + " " + fh.read()

    client = make_client(Predictor())
    resp = client.post(
        "/predict",
        json={
            "text": "baz",
            "num1": 5,
            "path": "data:text/plain;base64,"
            + base64.b64encode(b"wibble").decode("utf-8"),
        },
    )
    assert resp.status_code == 200
    assert resp.json() == {"output": "baz 50 wibble", "status": "success"}


def test_gt_lt():
    class Input(BaseModel):
        num: float = Field(..., gt=3, lt=10.5)

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> int:
            return input.num

    client = make_client(Predictor())
    resp = client.post("/predict", json={"num": 2})
    assert resp.status_code == 422

    resp = client.post("/predict", json={"num": 5})
    assert resp.status_code == 200


def test_options():
    class Options(Enum):
        foo = "foo"
        bar = "bar"

    class Input(BaseModel):
        text: Options

    class Predictor(cog.Predictor):
        def predict(self, input: Input) -> str:
            return str(input.text)

    client = make_client(Predictor())
    resp = client.post("/predict", json={"text": "foo"})
    assert resp.status_code == 200
    resp = client.post("/predict", json={"text": "baz", "num": 2})
    assert resp.status_code == 422