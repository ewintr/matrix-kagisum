FROM golang:1.20-bullseye as build
RUN apt update && apt install -y libolm3 libolm-dev

WORKDIR /src
COPY . ./
RUN go mod download
RUN go build -o /matrix-kagisum ./main.go

FROM debian:bullseye
RUN apt update && apt install -y libolm3 ca-certificates openssl

COPY --from=build /matrix-kagisum /matrix-kagisum
CMD /matrix-kagisum