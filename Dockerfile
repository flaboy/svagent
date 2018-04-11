FROM busybox:glibc
MAINTAINER wanglei <flaboy@shopex.cn>

RUN mkdir -p /data/
COPY svagent /data/
WORKDIR /data/
EXPOSE 6077
CMD ["/data/svagent host"]