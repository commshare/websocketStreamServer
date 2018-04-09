window.onload = function() {
    getLivingList();
}

function post(URL, PARAMS) {
    $.ajax({
        type: "POST",
        url: URL,
        data: PARAMS,
        success: function(data) {
            alert(data);
            var obj = JSON.parse(data);
            alert(obj.status);
            alert(obj.streams);
            for (var i = 0; i < obj.Streams.length; i++) {
                alert(obj.Streams[i]);
            }
        },
        err: function() {
            alert("failed");
        }
    })
}

function getLivingList() {
    post("/opt", { "op": 0 });
}