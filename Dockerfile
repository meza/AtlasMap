FROM golang:latest AS build
WORKDIR /out
COPY internal internal
COPY atlasmapserver atlasmapserver
COPY cmd cmd
COPY go.* ./

RUN go build -o /out/atlasmap cmd/atlasmap.go
FROM scratch AS bin
COPY --from=build /out/atlasmap /
ENTRYPOINT ./atlasmap
EXPOSE 3000