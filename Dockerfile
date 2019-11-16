FROM golang:1.13

EXPOSE 8080

WORKDIR /app

RUN mkdir /app

COPY . .

RUN go build -o twitterish .

ENTRYPOINT ["./twitterish"]
