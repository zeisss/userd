FROM ubuntu

RUN mkdir -p /opt
ADD ./userd /opt/userd

ADD ./start-userd /usr/local/bin/start-userd

EXPOSE 8080

CMD ["/usr/local/bin/start-userd"]