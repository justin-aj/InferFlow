GO ?= go
PYTHON ?= python

.PHONY: test router mock-backend loadgen-sample compose-up compose-down terraform-validate

test:
	$(GO) test ./...

router:
	$(GO) run ./cmd/router

mock-backend:
	$(GO) run ./cmd/mock-backend

loadgen-sample:
	$(PYTHON) loadgen/generator.py --requests 5 --output results/sample.csv

compose-up:
	docker compose up --build

compose-down:
	docker compose down --remove-orphans

terraform-validate:
	cd terraform/environments/dev && terraform init -backend=false && terraform validate
