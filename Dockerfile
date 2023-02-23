FROM ubuntu
WORKDIR /work
COPY pdfer.sh pdfer.sh
ENTRYPOINT [ "./pdfer.sh" ]