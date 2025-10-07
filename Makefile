SHELL = /bin/bash -o pipefail

.PHONY: test test-ci lint lint-ci fmt fmt-ci upgrade clean release lint-docs lint-docs-ci audit update-testdata encrypt decrypt sops

test:
	gotestsum --format pkgname --packages ./... -- -race -timeout 10m -cover -covermode atomic

test-ci:
	gotestsum --format pkgname --packages ./... --junitfile tests.xml -- -race -timeout 10m -coverprofile=coverage.txt -covermode atomic
	gocover-cobertura < coverage.txt > coverage.xml
	go tool cover -html=coverage.txt -o coverage.html

lint:
	golangci-lint run --output.text.colors --allow-parallel-runners --fix
	find testdata -name '*.go' -print0 | xargs -0 -n1 -I % golangci-lint run --output.text.colors --allow-parallel-runners --fix %

lint-ci:
	golangci-lint run --output.text.path=stdout --output.code-climate.path=codeclimate.json --issues-exit-code 0
	find testdata -name '*.go' -print0 | xargs -0 -n1 -I % golangci-lint run --output.text.path=stdout --output.code-climate.path=%_codeclimate.json --issues-exit-code 0 %
	jq -s 'add' codeclimate.json testdata/*_codeclimate.json > /tmp/codeclimate.json
	mv /tmp/codeclimate.json codeclimate.json
	rm -f testdata/*_codeclimate.json
	jq -e '. == []' codeclimate.json

fmt:
	go mod tidy
	git ls-files --cached --modified --other --exclude-standard -z | grep -z -Z '.go$$' | xargs -0 gofumpt -w
	git ls-files --cached --modified --other --exclude-standard -z | grep -z -Z '.go$$' | xargs -0 goimports -w -local gitlab.com/tozd/go/mediawiki

fmt-ci: fmt
	git diff --exit-code --color=always

upgrade:
	go run github.com/icholy/gomajor@v0.13.2 get all
	go mod tidy

clean:
	rm -rf coverage.* codeclimate.json testdata/*_codeclimate.json tests.xml coverage

release:
	npx --yes --package 'release-it@19.0.5' --package '@release-it/keep-a-changelog@7.0.0' -- release-it

lint-docs:
	npx --yes --package 'markdownlint-cli@~0.45.0' -- markdownlint --ignore-path .gitignore --ignore testdata/ --fix '**/*.md'

lint-docs-ci: lint-docs
	git diff --exit-code --color=always

audit:
	go list -json -deps ./... | nancy sleuth --skip-update-check

update-testdata:
	go -C testdata run update.go
	gzip --keep --force testdata/commons-testdata-mediainfo.json
	bzip2 --keep --force testdata/commons-testdata-mediainfo.json
	gzip --keep --force testdata/wikidata-testdata-all.json
	bzip2 --keep --force testdata/wikidata-testdata-all.json
	tar -C testdata --create --file testdata/enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar enwiki_namespace_0_0.ndjson
	rm -f testdata/enwiki_namespace_0_0.ndjson
	gzip --keep --force testdata/enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar
	bzip2 --keep --force testdata/enwiki-NS0-testdata-ENTERPRISE-HTML.json.tar

encrypt:
	gitlab-config sops --encrypt --mac-only-encrypted --in-place --encrypted-comment-regex sops:enc .gitlab-conf.yml

decrypt:
	SOPS_AGE_KEY_FILE=keys.txt gitlab-config sops --decrypt --in-place .gitlab-conf.yml

sops:
	SOPS_AGE_KEY_FILE=keys.txt gitlab-config sops .gitlab-conf.yml
