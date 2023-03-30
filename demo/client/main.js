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
          microphones {
            joined
            listenOnly
            talking
            muted
            voiceUserId
            __typename
          }
          cameras {
            streamId
            __typename
          }
          whiteboards {
            whiteboardId
            __typename
          }
          breakoutRoom {
            isDefaultName
            sequence
            shortName
            online
            __typename
          }
          __typename
        }
      }`;
      
      ws.send(JSON.stringify({id:"1", type: "start", payload:{ variables:{}, extensions: {}, query: query } }));
      
}



