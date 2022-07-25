const ID = document.querySelector('meta[name="ID"]').content;

// default hide username_password
document.getElementById("username_password").style.display = "none";
const selectType = document.getElementById("select_type");
selectType.onchange = function() {
  const selected = selectType.options[selectType.selectedIndex].value;
  console.log("selected ", selected);
  // show all first
  let show = [];
  let hide = [];

  if (selected === "Authorization Code Grant") {
    show = [
      "redirect_url_d",
      "client_id_d",
      "client_secret_d",
      "authorization_url_d",
      "token_url_d",
      "scope_d",
      "start_button",
      "code",
      "get_token_button",
      "token",
      "refresh_token_button",
    ];
    hide = ["username_password"];
  } else if (selected === "Implicit Grant") {
    show = [
      "redirect_url_d",
      "client_id_d",
      "authorization_url_d",
      "scope_d",
      "start_button",
      "code",
      "get_token_button",
      "token",
      "refresh_token_button",
    ];
    hide = ["token_url_d", "client_secret_d", "username_password"];

  } else if (selected === "Resource Owner Password Credentials Grant") {
    show = [
      "token_url_d",
      "scope_d",
      "client_id_d",
      "client_secret_d",
      "username_password",
      "token",
      "refresh_token_button",
    ];
    hide = [
      "authorization_url_d",
      "redirect_url_d",
      "code",
      "get_token_button",
    ];
  } else if (selected === "Client Credentials Grant") {
    show = [
      "token_url_d",
      "scope_d",

      "client_id_d",
      "client_secret_d",
    ];
    hide = [
      "username_password",
      "authorization_url_d",
      "redirect_url_d",
      "code",
      "get_token_button",
      "authorization_url_d",
      "redirect_url_d",
    ];
  }

  show.forEach((keyi) => {
    console.log("display", keyi, "block");
    document.getElementById(keyi).style.display = "block";
  });
  hide.forEach((keyi) => {
    console.log("display", keyi, "none");
    document.getElementById(keyi).style.display = "none";
  });
};

// set redirect_url
const redirectUrl = document.getElementById("redirect_url");
redirectUrl.value = location.protocol + "//" + location.host + location.pathname + "cb/" + ID;

function addLog(msg) {
  const logs = document.getElementById("logs");
  const logNode = document.createElement("code");
  logNode.className = "bg-secondary text-white p-2";
  logNode.innerText = msg;
  logs.appendChild(logNode);
}

// update expires_in
let xg_expires_in = 0;

function updateExpiresIn() {
  const nowts = Math.round(Date.now() / 1000);
  const secs = document.getElementById("expire_content").innerText;
  if (secs === undefined || secs === "") {
    return;
  }

  expires_in = xg_expires_in - nowts;
  if (expires_in <= 0) {
    document.getElementById("expire_content").innerText = "--";
  } else {
    document.getElementById("expire_content").innerText = expires_in.toString();
  }
}
window.setInterval(updateExpiresIn, 1000);

// websocket and logs

let wsurlSuffix = location.host + location.pathname + "ws/" + ID;
if (location.host === "quant67.com") {
  wsurlSuffix = "ws.quant67.com" + location.pathname + "ws/" + ID;
}
let wsurl = "wss://" + wsurlSuffix;
if (location.protocol === "http:") {
  wsurl = "ws://" + wsurlSuffix;
}

console.log("wsURL:", wsurl);
let ws = new WebSocket(wsurl);
ws.onopen = function(evt) {
  console.log("ws open");
};
ws.onmessage = function(evt) {
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
    addLog(data.data);
  } else if (data.type === "code") {
    document.getElementById("code_content").innerText = data.code;
  } else if (data.type === "token") {
    document.getElementById("access_token_content").innerText = data.access_token;
    if (data.refresh_token) {
      document.getElementById("refresh_token_content").innerText = data.refresh_token;
    }
    document.getElementById("expire_content").innerText = data.expires_in.toString();

    xg_expires_in = Math.round(Date.now() / 1000) + data.expires_in;
  }
};


ws.onclose = function(evt) {
  console.log("ws close");
};
ws.onerror = function(evt) {
  console.log("ws error");
};

let waitingNewWindowURL = true;
let newWindowURL = undefined;

// start link
const startButton = document.getElementById("start_button");
startButton.onclick = function() {
  const selected = selectType.options[selectType.selectedIndex].value;

  const eourl = document.getElementById("authorization_url").value;
  const clientID = document.getElementById("client_id").value;
  const clientSecret = document.getElementById("client_secret").value;
  const scope = document.getElementById("scope").value;
  const state = clientSecret;

  console.log("selected", selected);

  if (selected === "Authorization Code Grant") {
    const openURL = eourl + "?client_id=" + clientID + "&state=" +
          state + "&scope=" + scope + "&response_type=code&redirect_uri=" + redirectUrl.value;
    const newWindow = window.open(openURL, "authrorize", "resizable,scrollbars,status");

    logMSG = "opening the provider autorize url:\n" + openURL;
    addLog(logMSG);
  } else if (selected === "Implicit Grant") {
    const openURL = eourl + "?client_id=" + clientID + "&state=" +
          state + "&scope=" + scope + "&response_type=token&redirect_uri=" + redirectUrl.value;
    const newWindow = window.open(openURL, "authrorize", "resizable,scrollbars,status");
    const fn = (nw) => {
      if (nw.location.href === undefined) {
        console.log("new window closed");
        return;
      }
      let nurl = new URL(nw.location.href);
      let redurl = new URL(redirectUrl.value);
      if (waitingNewWindowURL && nurl.host === redurl.host && nurl.pathname === redurl.pathname) {
        newWindowURL = nw.location.href;
        waitingNewWindowURL = false;
        addLog("redirecting to\n" + newWindowURL);
        const kvps = nurl.hash.substring(1);
        kvps.split('&').forEach((kv) => {
          kvsp = kv.split('=');
          key = kvsp[0];
          value = kvsp[1];
          document.getElementById("expire_content").innerText = undefined;
          if (key === "access_token") {
            document.getElementById("access_token_content").innerText = value;
          } else if (key == "refresh_token") {
            document.getElementById("refresh_token_content").innerText = value;
          } else if (key === "expires_in") {
            xg_expires_in = Math.round(Date.now() / 1000) + parseInt(value);
            document.getElementById("expire_content").innerText = xg_expires_in.toString();
          }
        });
      }
      if (waitingNewWindowURL) {
        window.setTimeout(fn, 70, newWindow);
      }
    };
    window.setTimeout(fn, 70, newWindow);

    logMSG = "opening the provider autorize url:\n" + openURL;
    addLog(logMSG);


  } else if (selected === "Resource Owner Password Credentials Grant") {
    const selected = selectType.options[selectType.selectedIndex].value;
    let data = {
      "type": "get_token",
      "client_id": document.getElementById("client_id").value,
      "client_secret": document.getElementById("client_secret").value,
      "token_url": document.getElementById("token_url").value,
      "grant_type": "password",
      "username": document.getElementById("usernamex").value,
      "password": document.getElementById("passwordx").value,
      "scope": document.getElementById("scope").value,
    };

    ws.send(JSON.stringify(data));

    let logmsg = "get token from " + data.token_url + "\n";
    logmsg += "POST body to json:\n" + JSON.stringify(data);
    addLog(logmsg);
  } else if (selected === "Client Credentials Grant") {
    const selected = selectType.options[selectType.selectedIndex].value;
    let data = {
      "type": "get_token",
      "client_id": document.getElementById("client_id").value,
      "client_secret": document.getElementById("client_secret").value,
      "token_url": document.getElementById("token_url").value,
      "grant_type": "client_credentials",
      "scope": document.getElementById("scope").value,
    };

    ws.send(JSON.stringify(data));

    let logmsg = "get token from " + data.token_url + "\n";
    logmsg += "POST body to json:\n" + JSON.stringify(data);
    addLog(logmsg);
  }
};

// get token
const tokenButton = document.getElementById("get_token_button");
tokenButton.onclick = function() {
  console.log("get token button");
  const selected = selectType.options[selectType.selectedIndex].value;
  let data = {
    "type": "get_token",
    "code": document.getElementById("code_content").innerText,
    "client_id": document.getElementById("client_id").value,
    "client_secret": document.getElementById("client_secret").value,
    "token_url": document.getElementById("token_url").value,
    "redirect_uri": document.getElementById("redirect_url").value,
    "grant_type": "authorization_code"
  };

  if (selected == "Authorization Code Grant") {
    data.grant_type = "authorization_code";
  }

  ws.send(JSON.stringify(data));

  let logmsg = "get token from " + data.token_url + "\n";
  logmsg += "POST body to json:\n" + JSON.stringify(data);
  addLog(logmsg);
};

// refresh token
const refButton = document.getElementById("refresh_token_button");
refButton.onclick = function() {
  // const selected = selectType.options[selectType.selectedIndex].value;
  let data = {
    "type": "refresh_token",
    "refresh_token": document.getElementById("refresh_token_content").innerText,
    "client_id": document.getElementById("client_id").value,
    "client_secret": document.getElementById("client_secret").value,
    "token_url": document.getElementById("token_url").value,
    "grant_type": "refresh_token"
  };

  ws.send(JSON.stringify(data));

  let logmsg = "get token from " + data.token_url + "\n";
  logmsg += "POST body to json:\n" + JSON.stringify(data);
  addLog(logmsg);
};
