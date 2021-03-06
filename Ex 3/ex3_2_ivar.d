import std.stdio;
import std.getopt;
import std.string;
import std.socket;
import std.stream;
import std.socketstream;
import std.concurrency;

void main(string[] args) {
	debug(TCP)
		writefln("Debugging enabled");
	string ip = "127.0.0.1";
	string listen_ip = "127.0.0.1";
	ushort port = 124;
	ushort listen_port = 50000;
	getopt(args, 
		"host|h", &ip, 
		"port|p", &port, 
		"listen-port|lp", &listen_port, 
		"listen_ip|l", &listen_ip
	);
	debug(THREAD)
		writefln("Starting server");
	spawn(&tcp_server, listen_ip, listen_port);
	auto message = receiveOnly!bool();
	debug(THREAD)
		writefln("Server started");

	debug(THREAD)
		writefln("Starting client");
//	spawn(&tcp_connect, ip, port, listen_ip, listen_port);
	writefln("Main done");
}


void tcp_server(string listen_ip, ushort listen_port) {
	Socket listener = new TcpSocket;
	scope(exit) listener.close();
	assert(listener.isAlive);
//	listener.blocking = true;
	debug(TCP)
		writefln("Setting listen address and port");

	listener.bind(new InternetAddress(listen_ip, listen_port));
	listener.listen(5);
	debug(TCP)
		writefln("Entering socket while loop");
	debug(THREAD)
		writefln("SERVER: started ok");	
	ownerTid.send(true);
	Socket sock = listener.accept();
	assert(sock.isAlive);
	debug(TCP) 
		writefln("DEBUG: Incoming connection");
	while(true) {
		char[1024] buffer;
		sock.receive(buffer);
		writefln("From server: %s", buffer);
	}
	sock.close()
}

void tcp_connect(string ip, ushort port, string listen_ip, ushort listen_port) {
	debug(TCP)
		writefln("Connecting to host %s:%s", ip, port);
	Socket sock = new TcpSocket(new InternetAddress(ip, port));
	scope(exit) sock.close();
	
	debug(TCP)
		writefln("reading line");

	char[1024] line;
	sock.receive(line);
	writefln("string: %s", line);
	auto message = format("Connect to: %s:%s", listen_ip, listen_port);
	debug(TCP)
		writefln("Sending: %s", message);
	sock.send(message);

/*	
	for(int i=0; i < 5; i++) {
		char[] message;
		writefln("Write message to send:");
		readln(message);
		message[message.length-1] = '\0';
		debug(TCP)
			writefln("sending message");
		sock.send(message);
		debug(TCP)
			writefln("recivev\n");
		sock.receive(line);
		writefln("Recived line: %s", line);
	}
*/
}
	

