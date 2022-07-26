
const clientID = document.querySelector('meta[name="ID"]').content;

const wsurlSuffix = location.host + location.pathname + "ws/" + clientID;
let wsurl = "wss://" + wsurlSuffix;
if (location.protocol === "http:") {
  wsurl = "ws://" + wsurlSuffix;
}

console.log("wsURL:", wsurl);
const ws = new WebSocket(wsurl);
ws.onopen = function (evt) {
  console.log("ws open");
};
ws.onmessage = function (evt) {
  console.log("ws message:", evt.data);

  const data = JSON.parse(evt.data);
  if (data.type === "ping") {
	ws.send(JSON.stringify({
	  type: "pong",
	  data: ""
	}));
	return;
  }
  if (data.type === "log") {
	const logs = document.getElementById("logs");
	const logNode = document.createElement("code");
	logNode.className = "bg-secondary text-white p-2";
	logNode.innerText = data.data;
	logs.appendChild(logNode);
  }
}
ws.onclose = function (evt) {
  console.log("ws close");
};
ws.onerror = function (evt) {
  console.log("ws error");
};
