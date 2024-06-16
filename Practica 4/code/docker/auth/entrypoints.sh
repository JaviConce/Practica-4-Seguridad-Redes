#/bin/bash

iptables -P INPUT ACCEPT
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -p icmp -j ACCEPT

ip route del default
ip route add default via 10.0.2.2 dev eth0 


iptables -A INPUT -i eth0 -p tcp --dport 5000 -j ACCEPT #Borrar
iptables -A INPUT -s 10.0.1.4 -p tcp --dport 5000 -j ACCEPT #Broker
iptables -A INPUT -s 10.0.1.4 -p tcp --sport 5000 -j ACCEPT #Broker


service ssh start
service rsyslog start

./auth

if [ -z "$@" ]; then
    exec /bin/bash
else
    exec $@
fi