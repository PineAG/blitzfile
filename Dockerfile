FROM golang:alpine AS build
COPY . /app
WORKDIR /app
RUN go install
RUN go build blitzfile.go
USER 1000:1000

FROM alpine:latest
RUN mkdir /app
RUN mkdir /files
COPY --chown=1000:1000 --from=build /app/blitzfile /app/blitzfile
RUN chown -R 1000:1000 /app
RUN chown -R 1000:1000 /files
USER 1000:1000
ENV FILE_ROOT=/files
ENV PORT=8000
EXPOSE 3000
CMD ["/app/blitzfile"]