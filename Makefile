VENV?=.venv
PYTHON=$(VENV)/bin/python3
PIP=$(VENV)/bin/pip

.PHONY: venv
venv: $(VENV)/bin/activate
$(VENV)/bin/activate:
	python3 -m venv $(VENV)
	${PIP} install -r requirements.txt

.PHONY: clean
clean:
	rm -rf $(VENV)

.PHONY: run
run: venv
	${VENV}/bin/uvicorn main:app --reload

.PHONY: freeze
freeze: venv
	${PIP} freeze > requirements.txt
