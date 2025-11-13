FROM ubuntu:latest
LABEL authors="mymac"

ENTRYPOINT ["top", "-b"]