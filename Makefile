.PHONY: test fmt vet audit-sources clone-sources benchmark-matching

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clone-sources:
	./scripts/clone-sources.sh

audit-sources:
	./scripts/audit-sources.sh

benchmark-matching:
	go test -bench=. ./services/matching-engine/...

