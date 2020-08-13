FROM golang:1.14.4-alpine3.12

ARG VIPS_VERSION=8.9.2

ENV CGO_ENABLED=1
ENV GO111MODULE=on

RUN apk add --update \
    git bash curl \
    ca-certificates \
    wget \
    build-base \
    glib-dev \
    libxml2-dev \
    libjpeg-turbo-dev \
    libexif-dev \
    tiff-dev \
    libgsf-dev \
    libpng-dev \
    pango-dev \
  && wget https://github.com/jcupitt/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz \
  && tar -zxf vips-${VIPS_VERSION}.tar.gz \
  && cd vips-${VIPS_VERSION}/ \
  && ./configure \
    --prefix=/usr \
    --disable-debug \
    --disable-static \
    --disable-introspection \
    --disable-dependency-tracking \
    --enable-silent-rules \
    --without-python \
    --without-orc \
    --without-fftw \
  && make -s \
  && make install \
  && cd ../ \
  && rm -rf vips-${VIPS_VERSION}/ \
  && rm vips-${VIPS_VERSION}.tar.gz

RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build ./cmd/app/main.go

ENTRYPOINT ["./main", "-routes-file", "eskip/sample.eskip", "-verbose"]

