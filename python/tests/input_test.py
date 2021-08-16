import cog


def test_input():
    class Predictor(cog.Predictor):
        def predict(
            self,
            name: str = cog.Input(description="A string"),
            number: int = cog.Input(description="A number"),
        ):
            return name, number

    assert Predictor().get_type_signature() == {}
