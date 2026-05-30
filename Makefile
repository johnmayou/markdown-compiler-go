.PHONY: fmt
fmt:
	@go fmt

.PHONY: lint
lint:
	@bash ./scripts/ci-lint.sh

.PHONY: flint
flint: fmt lint

.PHONY: test
test:
	@bash ./scripts/ci-test.sh

.PHONY: release
release: ## tag and push to trigger release workflow (usage: make release VERSION=v0.1.0)
	@latest=$$(git tag --sort=-v:refname | head -1); \
	if [ -n "$$latest" ]; then \
		newer=$$(printf '%s\n' "$$latest" "$(VERSION)" | sort -V | tail -1); \
		if [ "$$newer" != "$(VERSION)" ]; then \
			echo "error: $(VERSION) is not greater than current latest $$latest"; exit 1; \
		fi; \
	fi
	@git tag $(VERSION) && git push origin $(VERSION)

.PHONY: release-dry-run
release-dry-run: ## build release bin locally without publishing
	goreleaser release --clean --snapshot