# build dependencies
FROM golang AS build
RUN mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

# copy source
COPY database database
COPY entities entities
COPY config config
COPY main.go main.go

# build the app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# run the app
FROM alpine
RUN mkdir /app
WORKDIR /app

COPY migrations migrations
COPY --from=build /app/heavens_throne heavens_throne

EXPOSE 443
ENTRYPOINT "./heavens_throne"
