var globalJobInfo = {
    id: null,
    status: null,
}

/**
 * getFormData serial data from form
 * @param {object} form - The jquery form object 
 * @returns {object} result - the serialized javascript object
 */
function getFormData($form) {
    var formArray = $form.serializeArray()
    var result = {}
    $.map(formArray, function (n, i) {
        result[n["name"]] = n["value"]
        console.log(n)
    })
    return result
}


/**
 * getFormData serial data from form
 * @param {object} form - The jquery form object 
 * @returns {object} result - the serialized javascript object
 */
function validateConfig(result) {
    if (result["server"] === "") {
        $("input[name=server]").addClass("error-input")
        toastr.error('dns server is empty', 'Config Error')
        return false
    }
    if (result["domain"] === "") {
        result["domain"] = "."
    }
    var port = parseInt(result["port"])
    if(isNaN(port)){
        result["port"] = "53"
    }
    if (port <= 0 || port > 65535) {
        toastr.error('Port number should be in [0-65535]', 'Port Error')
        return false
    }
    result["qps"] = isNaN(parseInt(result["qps"])) ? 100 : parseInt(result["qps"])
    if (result["qps"] <= 0) {
        toastr.error('QPS number should be larger than 0', 'QPS Error')
        return false
    }
    result["domain_random_length"] = isNaN(parseInt(result["domain_random_length"])) ? 5 : parseInt(result["domain_random_length"])
    if (result["domain_random_length"] < 0) {
        toastr.error('Random length number should not smaller than 0', 'Length Error')
        return false
    }
    if (result["duration"] === "") {
        result["duration"] = "60s"
    }
    return true
}

function Logger(id) {
    this.messageArray = [];
    this.logBoxContainer = $("#" + id)
    if (this.logBoxContainer.length === 0) {
        console.error("Logger init fail: id not exist")
        return
    }
    var self = this;
    this.timer = setInterval(function () {
        // 定期执行数据清理工作,仅仅保留其中的50条数据
        // 清理数组中的数据
        if (self.messageArray.length > 100) {
            self.messageArray.splice(0, self.messageArray.length - 50)
        }
        // 清理DOM中的元素,只保留50个最新的元素
        if ($(".message").length > 100) {
            $(".message").splice(50, $(".message").length).map(function (div) {
                div.remove();
            })
        }
    }, 10000)
}

/**
 * 获取当前的时间
 * @description 输出信息 "01:02:23"
 *
 */
function getDate() {
    var d = "",
        s = "",
        t = "";
    d = new Date();
    t = d.getHours();
    s += (t > 9 ? "" : "0") + t + ":";
    t = d.getMinutes();
    s += (t > 9 ? "" : "0") + t + ":";
    t = d.getSeconds();
    s += (t > 9 ? "" : "0") + t;
    return s;
}

/**
 * 接收消息并格式化信息
 *
 * @param {object} message 代表了信息的对象内容
 * @property {string} status    - 代表了信息的状态
 * @property {string} message   - 代表了信息的详细内容
 */
Logger.prototype.formatMessage = function (message) {
    var status = ""
    switch (message.status) {
        case "error":
            status = '<p class="message error">' + getDate() + " [Error] " + message.message + '</p>'
            break;
        case "warning":
            status = '<p class="message warning">' + getDate() + " [Warning] " + message.message + '</p>'
            break;
        default:
            status = '<p class="message info">' + getDate() + " [Info] " + message.message + '</p>'
    }
    return status
}

/**
 * 追加消息并增加到DOM
 *
 * @param {string} status    - 代表了信息的状态
 * @param {string} message   - 代表了信息的详细内容
 */
Logger.prototype.appendMessage = function (message, status) {
    if (typeof status === 'undefined') {
        status = "info"
    }
    var messageStruct = {
        message: message,
        status: status
    }
    this.messageArray.push(messageStruct)
    this.logBoxContainer.prepend(this.formatMessage(messageStruct))
}
/**
 * 记录一般性的通用消息并增加到DOM
 *
 * @param {string} message   - 代表了信息的详细内容
 */
Logger.prototype.info = function (message) {
    this.appendMessage(message, "info")
}
/**
 * 记录错误消息并增加到DOM
 *
 * @param {string} message   - 代表了信息的详细内容
 */
Logger.prototype.error = function (message) {
    this.appendMessage(message, "error")
}
/**
 * 记录普通告警消息并增加到DOM
 *
 * @param {string} message   - 代表了信息的详细内容
 */
Logger.prototype.warning = function (message) {
    this.appendMessage(message, "warning")
}
Logger.prototype.batch = function (messages) {
    for (var i = 0; i < messages.length; i++) {
        switch (messages[i]['level']) {
            case "info":
                this.info(messages[i]['msg'])
                break;
            case "warn":
                this.warning(messages[i]['msg'])
                break;
            case "error":
                this.error(messages[i]['msg'])
                break;
            default:
                this.info(messages[i]['msg'])
        }
    }
}
var logger = new Logger("console-info")




$(document).ready(function () {
    var historyTable = $('#history-table').DataTable( {
        "processing": true,
        "paging": true,
        "serverSide": true,
        "ordering": false,
        "lengthChange": false,
        "info": false,
        "ajax": "/history",
        "dataSrc": "data",
        "pagingType": "simple",
        "columns": [
            {data: "server"},
            {data: "port"},
            {data: "duration"},
            {data: "qps"},
            {data: "domain"},
            {data: "domain_random_length"},
            {data: "query_type"},
            {
                "targets": -2,
                "data": function(row){
                    return moment(row.CreatedAt).fromNow();
                },
            },
            {
                "targets": -1,
                "data": null,
                "defaultContent": "<button>Reload</button>"
            } 
        ]
    });
    $(".config-submit").click(function () {
        var result = getFormData($('form[name="config"]'))
        debugger
        if (validateConfig(result) === false) {
            return
        }
        debugger
        $.ajax({
            type: "POST",
            url: "/start",
            data: JSON.stringify(result),
            success: function (response) {
                $('.master-running').removeClass("hide")
                globalJobInfo.id = response["id"]
                historyTable.ajax.reload();
            },
            error: function (err) {
                console.log(err)
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error, "Error")
                } else {
                    toastr.error("Error", "Server Fail")
                }
            },
            contentType: "application/json"
        })
    })
    $("#delete-agent").click(function () {
        var ipWithPort = $(this).attr("data-item")
        var data = {
            "ipaddress": ipWithPort.split(":")[0],
            "port": ipWithPort.split(":")[1]
        }
        $.ajax({
            type: "DELETE",
            url: "/nodes",
            data: JSON.stringify(data),
            success: function (data) {
                toastr.info("delete success")
                window.location.reload()
            },
            error: function (err) {
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error,"delete fail")
                } else {
                    toastr.error("delete fail","Sever Fail")
                }
            },
            contentType: "application/json"
        })
    })
    function updateAgentEnableStatus(ip,port,enable){
        var data = {
            "ipaddress": ip,
            "port": port,
            "enable":enable
        }
        $.ajax({
            type: "POST",
            url: "/update-node",
            data: JSON.stringify(data),
            success: function (data) {
                toastr.info("update success")
                window.location.reload()
            },
            error: function (err) {
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error,"update fail")
                } else {
                    toastr.error("update fail","Sever Fail")
                }
            },
            contentType: "application/json"
        })
    }
    $("#disable-agent").click(function () {
        var ipWithPort = $(this).attr("data-item").split(":")

        updateAgentEnableStatus(ipWithPort[0],ipWithPort[1],false)
    })
    $("#enable-agent").click(function(){
        var ipWithPort = $(this).attr("data-item").split(":")
        updateAgentEnableStatus(ipWithPort[0],ipWithPort[1],true)
    })
    $("#show-history").click(function () {
        var isHide = $(".history-box").hasClass("hide")
        if(isHide === true){
            $(".history-box").removeClass("hide")
        }else{
            $(".history-box").addClass("hide")
        }
    })
   

    function updateConfigurationFromData(data){
        var keys = Object.keys(data)
        console.log(data)
        for(var i = 0; i<keys.length; i++){
            var inputSelector = "form[name='config'] input[name='"+keys[i]+"']"
            if($(inputSelector).length===1){
                    $(inputSelector).val(data[keys[i]])
                    continue
            }
            if($(inputSelector).length===2 && $(inputSelector).is(':radio')){
                if(data[keys[i]]==="true"){
                    $(inputSelector).eq(0).attr("checked","checked")
                }else{
                    $(inputSelector).eq(1).attr("checked","checked")
                }
            }
        }
    }

    $('#history-table tbody').on( 'click', 'button', function () {
        var data = historyTable.row( $(this).parents('tr') ).data();
        // hide the table 
        $(".history-box").addClass("hide")
        // set to current configuration
        updateConfigurationFromData(data)
        
    } );
    function updatePingStatus(ipinfo, pingSuccess){
        if(pingSuccess === true){
            $(".agent-ping[data-item='"+ipinfo+"']").find("i").removeClass("hide")
        }else{
            $(".agent-ping[data-item='"+ipinfo+"']").find("i").addClass("hide")
        }    
    }

    $(".small-ping-button").click(function () {
        var ipWithPort = $(this).attr("data-item")
        var data = {
            "ipaddress": ipWithPort.split(":")[0],
            "port": ipWithPort.split(":")[1]
        }
        $.ajax({
            type: "POST",
            url: "/ping",
            data: JSON.stringify(data),
            success: function (data) {
                toastr.success("ping success")
                updatePingStatus(ipWithPort,true)
            },
            error: function (err) {
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error,"ping fail")
                } else {
                    toastr.error("ping fail","Sever Fail")
                }
                updatePingStatus(ipWithPort,false)
            },
            contentType: "application/json"
        })
    })


    $('.config-kill').click(function () {
        $.ajax({
            type: "GET",
            url: "/stop",
            success: function (response) {
                $('.master-running').addClass("hide")
                toastr.success("stop traffic success")
            },
            error: function (err) {
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error, "Error")
                } else {
                    toastr.error("ServerFail")
                }
            },
            contentType: "application/json"
        })
    })

    $('.new-agent').click(function () {
        var data = getFormData($("form[name='new-agent']"))
        if (typeof data.ipaddress === 'undefined' || data.ipaddress === "") {
            toastr.error('IP address does not exist', 'IP Error')
            return
        }
        if(data.port === ""){
            data.port = "8998"
        }
        if (parseInt(data.port) < 0  || parseInt(data.port) > 65535) {
            toastr.error('Port number should be in [0-65535]', 'Port Error')
            return
        }
        $(".add-node-loading").removeClass("hide")
        $.ajax({
            type: "POST",
            url: "/nodes",
            data: JSON.stringify(data),
            success: function (response) {
                $(".add-node-loading").addClass("hide")
                window.location.reload()
            },
            error: function (err) {
                $(".add-node-loading").addClass("hide")
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error,"Add new node fail")
                } else {
                    toastr.error("Add new node fail", "ServerFail")
                }
            },
            contentType: "application/json"
        })
    })
    function updateAgentStatus(nodes){
        nodes.map(function(node){
            var nodeInfo = node.ip + ":" + node.port;
            if(node.status === "running"){
                $(".agent-running[data-item='"+nodeInfo+"']").find("i.running-success").removeClass("hide")
            }else{
                $(".agent-running[data-item='"+nodeInfo+"']").find("i.running-success").addClass("hide")
            }
        })
    }
    // 每隔2秒发送一次查询日志的请求
    var queryStatusTimer = setInterval(function () {
        $.ajax({
            type: "GET",
            url: "/status",
            success: function (response) {
                    if(response.id){
                        globalJobInfo.id = response.id
                    }
                    if (response.status ) {
                        switch(response.status){
                            case "running":
                            case "init":
                            case "stopping":
                                $(".config-submit").attr("disabled", true)
                                $(".master-running").removeClass("hide")
                                break;
                            case "stopped":
                            default:
                                $(".config-submit").attr("disabled", false)
                                $(".master-running").addClass("hide")
                                break;
                        }
                    }
                    if(response.messages && response.messages.length > 0){
                        logger.batch(response.messages)
                    }
                    if(response.nodes && response.nodes.length > 0){
                        updateAgentStatus(response.nodes)
                    }
            },
            error: function(err){
                if (err && err.responseJSON && err.responseJSON.error) {
                    toastr.error(err.responseJSON.error, "Error")
                } else {
                    toastr.error("Error", "Server Fail")
                    // clearInterval(queryStatusTimer)
                }
            },
            contentType: "application/json"
        })
    }, 2000)
})