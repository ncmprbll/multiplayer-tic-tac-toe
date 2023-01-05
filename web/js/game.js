var id = window.location.href.split("/").pop() || ""

if (id !== "") {
    const ACTION_MOVE = "move";
    const ACTION_UPDATE = "update";
    const ACTION_STATE_UPDATE = "state_update";
    const ACTION_CHAT = "chat";
    const ACTION_SWITCH = "switch";
    const ACTION_ROUND_END = "round_end";

    const FIELD_NOT_SET = 0;
    const FIELD_X = 1;
    const FIELD_O = 2;

    const GAME_NOT_STARTED = 0;
    const GAME_WAITING_FOR_X = 1;
    const GAME_WAITING_FOR_O = 2;
    const GAME_ROUND_END = 3;
    const GAME_OVER = 4;

    function vToXY(v) {
        return {x: (v - 1) % 3, y: Math.floor((v - 1) / 3)}
    }

    function XYTov(x, y) {
        return x + y * 3 + 1
    }

    const getGameCookieValue = (name) => (
        document.cookie.match('(^|;)\\s*' + id + "_" + name + '\\s*=\\s*([^;]+)')?.pop() || ''
    )

    var socket = new WebSocket("ws://" + location.host + "/ws/" + id + "/?id=" + getGameCookieValue("id"));

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

        if (issystem) {
            div.append(date, t);
        } else {
            div.append(date, s, t);
        }

        chat.append(div);
    }

    socket.addEventListener("message", function (event) {
        var data = JSON.parse(event.data);
        if (data.action === ACTION_MOVE) {
            const button = document.getElementById(XYTov(data.x, data.y));
            if (data.value == FIELD_X) {
                button.innerText = "X";
            } else if (data.value == FIELD_O) {
                button.innerText = "O";
            }
        } else if (data.action === ACTION_UPDATE) {
            for (var i = 1; i <= 9; i++) {
                var {x, y} = vToXY(i)
                const button = document.getElementById(i);

                if (data.value[x][y] === FIELD_X) {
                    button.innerText = "X";
                } else if (data.value[x][y] === FIELD_O) {
                    button.innerText = "O";
                } else if (data.value[x][y] === FIELD_NOT_SET) {
                    button.innerText = "";
                }

                button.style.backgroundColor = null;
            }
        } else if (data.action === ACTION_STATE_UPDATE) {
            const infobox = document.getElementById("infobox");
            var text = "...";

            const whoami = getGameCookieValue("whoami");

            switch (data.value) {
                case GAME_NOT_STARTED:
                    text = "Waiting";
                    break;
                case GAME_WAITING_FOR_X:
                    if (whoami === "X") {
                        text = "Your move";
                    } else if (whoami == "O") {
                        text = "Opponent's move"
                    } else {
                        text = "X's move"
                    }

                    break;
                case GAME_WAITING_FOR_O:
                    if (whoami === "X") {
                        text = "Opponent's move";
                    } else if (whoami == "O") {
                        text = "Your move"
                    } else {
                        text = "O's move"
                    }

                    break;
                case GAME_ROUND_END:
                    text = "Round is over";
                    break;  
                case GAME_OVER:
                    text = "Game ended";
                    break;
                default:
                    text = "Unknown state";
            }

            infobox.innerHTML = text;
        } else if (data.action === ACTION_CHAT) {
            createMessage(data.sender, data.text, data.timestamp, data.issystem);
        } else if (data.action === ACTION_SWITCH) {
            const whoami = getGameCookieValue("whoami");

            if (whoami === "X") {
                document.cookie = id + "_whoami=O"
            } else if (whoami === "O") {
                document.cookie = id + "_whoami=X"
            }
        } else if (data.action === ACTION_ROUND_END) {
            for (const i in data.value) {
                const field = data.value[i];
                const button = document.getElementById(XYTov(field[0], field[1]));

                button.style.backgroundColor = "#5566cd";
            }
        }
    });

    function click() {
        var {x, y} = vToXY(this.id)

        socket.send(JSON.stringify({player: getGameCookieValue("id"), action: ACTION_MOVE, x: x, y: y}));
    }

    for (var i = 1; i <= 9; i++) {
        document.getElementById(i).addEventListener("click", click);
    }

    const textarea = document.getElementById("chat-textarea");
    const send = document.getElementById("send");

    function handler(event) {
        if (event.type === "keypress" && event.key !== "Enter") {
            return;
        }

        const text = textarea.value.trim();

        if (text !== "") {
            socket.send(JSON.stringify({player: getGameCookieValue("id"), action: ACTION_CHAT, text: text}));
        }

        textarea.value = "";
    }

    textarea.addEventListener("keypress", handler);
    send.addEventListener("click", handler);
}