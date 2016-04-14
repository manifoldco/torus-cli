var net = require('net');

function main () {
  var SOCKET_PATH = process.env.SOCKET_PATH;
  if (!SOCKET_PATH) {
    throw new Error('SOCKET_PATH env var must be set');
  }

  var socket = new net.Socket();
  socket.on('connect', function () {
    setInterval(function() {
      socket.write(JSON.stringify({"type":"request", id: 'hello', command: 'set', body: { value: 'woo '} }));
      socket.write(JSON.stringify({"type":"request", id: 'woohoo', command: 'get' }));
    }, 2000);
  });

  socket.on('data', function(buf) {
    console.log('got data', buf.toString('utf-8'));
  });

  socket.on('timeout', function() {
    console.log('Socket timeout');
    process.exit(1);
  });

  socket.on('error', function (err) {
    console.log('Error: ', (err.stack) ? err.stack : err);
    process.exit(1);
  });

  socket.on('close', function () {
    console.log('Socket closed');
    console.log(arguments);
    process.exit(1);
  });

  socket.connect({ path: SOCKET_PATH });
}

if (require.main === module) {
  main();
}
