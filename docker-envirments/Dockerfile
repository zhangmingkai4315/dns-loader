FROM  sameersbn/bind:9.9.5-20170626
ADD ./root-conf/* /etc/bind/
EXPOSE 53/udp
CMD ["named", "-g"]