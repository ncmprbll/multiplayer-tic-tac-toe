var id = window.location.href.split("/").pop() || ""
var socket

function vToXY(v) {
    return {x: (v - 1) % 3, y: Math.floor((v - 1) / 3)}
}

function XYTov(x, y) {
    return x + y * 3 + 1
}

const getCookieValue = (name) => (
    document.cookie.match('(^|;)\\s*' + name + '\\s*=\\s*([^;]+)')?.pop() || ''
)

if (id !== "") {
    function createMessage(sender, text, timestamp, issystem) {
        const chat = document.getElementById("chat");

        const div = document.createElement("div");
        div.classList.add("message");
    
        const date = document.createElement("span");
        date.classList.add("message-date");

        const s = document.createElement("span");
        s.classList.add("message-sender");

        const t = document.createElement("span");
        t.classList.add("message-text");

        date.innerText = timestamp;
        s.innerText = sender;
        t.innerText = text;

        div.append(date, s, t);
        chat.append(div);
    }

    socket = new WebSocket("ws://" + location.host + "/ws/" + id);

    socket.addEventListener("message", function (event) {
        var data = JSON.parse(event.data);
        if (data.action === "move") {
            const button = document.getElementById(XYTov(data.x, data.y));
            if (data.value == 1) {
                button.innerText = "X";
            } else if (data.value == 2) {
                button.innerText = "O";
            }
        } else if (data.action == "update") {
            for (var i = 1; i <= 9; i++) {
                var {x, y} = vToXY(i)
                const button = document.getElementById(i);

                if (data.value[x][y] == 1) {
                    button.innerText = "X";
                } else if (data.value[x][y] == 2) {
                    button.innerText = "O";
                }
            }
        } else if (data.action == "state_update") {
            const infobox = document.getElementById("infobox");
            var text = "...";

            switch (data.value) {
                case 0:
                    text = "Waiting";
                    break;
                case 1:
                    text = "X's move";
                    break;
                case 2:
                    text = "O's move";
                    break;
                case 3:
                    text = "Game ended";
                    break;
                default:
                    text = "Unknown state";
            }

            infobox.innerHTML = text;
        } else if (data.action == "chat") {
            createMessage(data.sender, data.text, data.timestamp, data.issystem);
        }
    });

    function click() {
        var {x, y} = vToXY(this.id)

        socket.send(JSON.stringify({player: getCookieValue("player-id"), action: "move", x: x, y: y}));
    }

    for (var i = 1; i <= 9; i++) {
        document.getElementById(i).addEventListener("click", click);
    }

    const textarea = document.getElementById("chat-textarea");
    const send = document.getElementById("send");

    function handler(event) {
        if (event.type === "keypress" && event.key != "Enter") {
            return;
        }

        const text = textarea.value.trim();

        socket.send(JSON.stringify({player: getCookieValue("player-id"), action: "chat", text: text}));
        textarea.value = "";
    }

    textarea.addEventListener("keypress", handler);
    send.addEventListener("click", handler);
}