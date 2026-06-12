FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY generator/ generator/
COPY resume.yaml ./
RUN go run ./generator -in resume.yaml -out /index.html

FROM nginx
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /index.html /etc/nginx/html/index.html
