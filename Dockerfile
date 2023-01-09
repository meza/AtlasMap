FROM scratch
COPY dist/atlasmap /
ENTRYPOINT /atlasmap
EXPOSE 3000