FROM debian
RUN apt update && apt install -y ca-certificates
WORKDIR /fetcher
COPY logfetcher .
CMD ["./logfetcher"]