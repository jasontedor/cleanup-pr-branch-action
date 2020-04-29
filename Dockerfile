FROM golang:1.12.5-alpine3.9

WORKDIR /go/src/cleanup-pr-branch-action
COPY cleanup-pr-branch-action.go .

# git needed for go get
RUN apk --no-cache add git

RUN CGO_ENABLED=0 GOOS=linux go get github.com/google/go-github/github \
  && go get golang.org/x/oauth2 \
  && go install

FROM alpine:3.9
RUN mkdir /cleanup-pr-branch-action
WORKDIR /cleanup-pr-branch-action
COPY --from=0 /go/bin/cleanup-pr-branch-action .

ENTRYPOINT ["cleanup-pr-branch-action"]