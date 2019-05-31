FROM  sameersbn/bind:9.9.5-20170626
ADD ./conf/* /etc/bind/
EXPOSE 53/udp
EXPOSE 53/tcp
CMD ["named", "-g"]