var id = window.location.href.split("/").pop() || ""
var socket

function vToXY(v) {
    return {x: (v - 1) % 3, y: Math.floor((v - 1) / 3)}
}

function XYTov(x, y) {
    return x + y * 3 + 1
}

if (id !== "") {
    socket = new WebSocket("ws://localhost:1337/ws/" + id);

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
        }
    });

    const getCookieValue = (name) => (
        document.cookie.match('(^|;)\\s*' + name + '\\s*=\\s*([^;]+)')?.pop() || ''
    )

    function click(event) {
        var {x, y} = vToXY(this.id)

        socket.send(JSON.stringify({player: getCookieValue("player-id"), action: "move", x: x, y: y}));
    }

    for (var i = 1; i <= 9; i++) {
        document.getElementById(i).addEventListener("click", click);
    }
}