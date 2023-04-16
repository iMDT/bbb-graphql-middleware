const WebSocket = require('ws');
//const ws = new WebSocket("wss://bbb-graphql-test-server.bbb.imdt.dev/v1/graphql", ['graphql-ws'], {
const ws = new WebSocket("ws://127.0.0.1:8378/v1/graphql", ['graphql-ws'], {
    headers: {
        "Cookie": "JSESSIONID=C12A18A5E3FDC88EB8B84B16CA59308F; sessionID="
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

    ws.send(`{"type":"connection_init","payload":{"headers":{"X-Session-Token":"c2ne5fnsc1jtaeiz"}}}	`);

    const query = `subscription {
        user(where: {joined: {_eq: true}}, order_by: {name: asc}) {
          userId
          name
          role
          color
          avatar
          emoji
          avatar
          presenter
          pinned
          locked
          authed
          mobile
          clientType
          leftFlag
          loggedOut
          __typename
        }
      }`;
      
      const payload = { variables:{}, extensions: {}, query: query };
    //   console.log(`Sending: ${JSON.stringify(payload)}`);
      ws.send(JSON.stringify({id:"1", type: "start", payload }));
      
}



