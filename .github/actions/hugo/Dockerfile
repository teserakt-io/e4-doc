FROM golang:1.12

RUN git clone https://github.com/gohugoio/hugo.git \
    && cd hugo \
    && go install --tags extended

ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
