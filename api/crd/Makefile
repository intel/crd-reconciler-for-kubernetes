all: validate

validate:
	# Positive tests
	jsonschema -i examples/stream-prediction-job-valid-1.json stream-prediction-job-spec.json

	# Negative tests
	(jsonschema -i examples/stream-prediction-job-invalid-1.json stream-prediction-job-spec.json 2> /dev/null ; test $$? -eq 1)