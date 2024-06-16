#!/bin/bash

echo 1 > /proc/sys/net/ipv4/ip_forward

iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

iptables -A INPUT -i lo -j ACCEPT #Permitir Ping

iptables -A INPUT -i eth0 ACCEPT
iptables -A INPUT -i eth1 ACCEPT 
iptables -A INPUT -i eth3 ACCEPT

iptables -A FORWARD -i eth0 -j ACCEPT
iptables -A FORWARD -i eth1 -j ACCEPT
iptables -A FORWARD -i eth2 -j ACCEPT
iptables -A FORWARD -i eth3 -j ACCEPT

iptables -A INPUT -p icmp -j ACCEPT
iptables -A FORWARD -p icmp -j ACCEPT
iptables -t nat -A POSTROUTING -o eth0 -p icmp -j MASQUERADE

#Permitir protocolo TCP y UDP
iptables -A INPUT -p udp --sport 53 -j ACCEPT
iptables -A INPUT -p ucp --dport 53 -j ACCEPT
iptables -A INPUT -p tcp --sport 53 -j ACCEPT
iptables -A INPUT -p tdp --dport 53 -j ACCEPT

iptables -A FORWARD -p udp --sport 53 -j ACCEPT
iptables -A FORWARD -p ucp --dport 53 -j ACCEPT
iptables -A FORWARD -p tcp --sport 53 -j ACCEPT
iptables -A FORWARD -p tdp --dport 53 -j ACCEPT

iptables -t nat -A POSTROUTING -o eth0 -p udp --dport 53 -j MASQUERADE
iptables -t nat -A POSTROUTING -o eth0 -p ucp --sport 53 -j MASQUERADE
iptables -t nat -A POSTROUTING -o eth0 -p tcp --dport 53 -j MASQUERADE
iptables -t nat -A POSTROUTING -o eth0 -p tdp --sport 53 -j MASQUERADE

#Aceptamos TCP en el puerto 443
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --sport 443 -j ACCEPT

#Reenvio de tcp en el puerto 443
iptables -A FORWARD -p tcp --dport 443 -j ACCEPT 
iptables -A FORWARD -p tcp --sport 443 -j ACCEPT

iptables -t nat -A POSTROUTING -o eth0 -p tcp --dport 443 -j MASQUERADE
iptables -t nat -A POSTROUTING -o eth0 -p tcp --sport 443 -j MASQUERADE

#Reenvio de la interfaz de entrada a la interfaz de salida
iptables -A FORWARD -i eth1 -o eth3 -j ACCEPT
iptables -A FORWARD -i eth3 -o eth1 -j ACCEPT
iptables -A FORWARD -i eth2 -o eth2 -j ACCEPT

iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 5000 -j DNAT --to-destination 10.0.1.4:5000
iptables -t nat -A POSTROUTING -o eth1 -p tcp --dport 5000 -s 172.17.0.0/16 -d 10.0.1.4 -j SNAT --to-source 10.0.1.2

#Reenvio de la interfaz de entrada y salida por TCP en el puerto 5000
iptables -A FORWARD -i eth0 -o eth1 -p tcp --syn --dport 5000 -m state --state NEW -j ACCEPT
iptables -A FORWARD -i eth1 -o eth0 -p tcp --syn --sport 5000 -m state --state NEW -j ACCEPT

iptables -A FORWARD -i eth0 -o eth1 -p tcp --syn --dport 22 -m state --state NEW -j ACCEPT
iptables -A FORWARD -i eth1 -o eth3 -p tcp --syn --dport 22 -m state --state NEW -j ACCEPT
iptables -A FORWARD -i eth0 -o eth1 -p tcp --syn --dport 22 -m state --state NEW -j ACCEPT
iptables -A FORWARD -i eth1 -o eth2 -p tcp --syn --dport 22 -m state --state NEW -j ACCEPT

iptables -A FORWARD -i eth1 -o eth3 -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -i eth3 -o eth1 -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -i eth0 -o eth1 -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A FORWARD -i eth1 -o eth0 -m state --state ESTABLISHED,RELATED -j ACCEPT

iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 22 -j DNAT --to-destination 10.0.1.3
iptables -t nat -A POSTROUTING -o eth1 -p tcp --dport 22 -s 172.17.0.0/16 -d 10.0.1.3 -j SNAT --to-source 10.0.1.2



iptables -A FORWARD -i eth1 -o eth2 -p tcp --dport 22 -j ACCEPT
iptables -A FORWARD -i eth2 -o eth1 -p tcp --sport 22 -j ACCEPT
iptables -A FORWARD -i eth2 -o eth1 -p tcp --dport 22 -j ACCEPT
iptables -A FORWARD -i eth1 -o eth2 -p tcp --sport 22 -j ACCEPT
iptables -A FORWARD -i eth3 -o eth1 -p tcp --dport 22 -j ACCEPT
iptables -A FORWARD -i eth1 -o eth3 -p tcp --sport 22 -j ACCEPT

iptables -A INPUT -p tcp --dport 22 -i eth2 -s 10.0.3.3 -j ACCEPT

service ssh start
rsyslogd

if [ -z "$@" ]; then
    exec /bin/bash
else
    exec $@
fi
