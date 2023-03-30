const WebSocket = require('ws');
const ws = new WebSocket("ws://127.0.0.1:8378/v1/graphql", ['graphql-ws'], {
    headers: {
        "Cookie": "JSESSIONID=B09E8B79DB7B0B05C38322AE9C5BFCE3; sessionID="
    }
});
 
ws.onmessage = (event) => {
    console.log(`Received: ${event.data}`);
}

ws.onclose = (event) => {
    console.log(`Closed: ${event.reason}`);
    process.exit(0);
}

ws.onopen = (event) => {
    const num = new Date().getTime();
    let msg = 0;

    ws.send(`{"type":"connection_init","payload":{"headers":{"X-Session-Token":"ezhyvo77prgcasbd"}}}`);
}



