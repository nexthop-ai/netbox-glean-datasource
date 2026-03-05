FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /netbox-glean-datasource .

FROM gcr.io/distroless/static-debian12
COPY --from=builder /netbox-glean-datasource /netbox-glean-datasource
ENTRYPOINT ["/netbox-glean-datasource"]
CMD ["serve"]
