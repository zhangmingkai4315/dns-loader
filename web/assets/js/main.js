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
    $(".config-submit").click(function () {
        var result = getFormData($('form[name="config"]'))
        if (validateConfig(result) === false) {
            return
        }
        $.ajax({
            type: "POST",
            url: "/start",
            data: JSON.stringify(result),
            success: function (data) {
                var jsonObj = data.responseJSON
                $('.master-running').removeClass("hide")
                globalJobInfo.id = jsonObj["id"]
            },
            error: function (data) {
                var jsonObj = data.responseJSON
                if (jsonObj && jsonObj['status'] && jsonObj['status'] === 'error') {
                    toastr.error(jsonObj['message'] || 'service not available')
                    return
                }
            },
            contentType: "application/json"
        })
    })
    $(".small-delete-button").click(function () {
        var ipWithPort = $(this).attr("data-item")
        var data = {
            "ipaddress": ipWithPort.split(":")[0],
            "port": parseInt(ipWithPort.split(":")[1])
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
                if (err && err.responseJSON && err.responseJSON.message) {
                    toastr.error("Delete fail", err.responseJSON.message)
                } else {
                    toastr.error("Delete fail")
                }
            },
            contentType: "application/json"
        })
    })

    $(".small-ping-button").click(function () {
        var ipWithPort = $(this).attr("data-item")
        var data = {
            "ipaddress": ipWithPort.split(":")[0],
            "port": parseInt(ipWithPort.split(":")[1])
        }
        $.ajax({
            type: "POST",
            url: "/ping",
            data: JSON.stringify(data),
            success: function (data) {
                toastr.success("ping success")
            },
            error: function (err) {
                if (err && err.responseJSON && err.responseJSON.message) {
                    toastr.error("ping fail", err.responseJSON.message)
                } else {
                    toastr.error("ping fail")
                    // change the color of status to black
                }
            },
            contentType: "application/json"
        })
    })


    $('.config-kill').click(function () {
        console.log("stop signal send to master server")
        $.ajax({
            type: "GET",
            url: "/stop",
            success: function (response) {
                console.log(response)
                $('.master-running').addClass("hide")
                toastr.success("stop traffic success")
            },
            error: function (err) {
                if (err && err.responseJSON && err.responseJSON.message) {
                    toastr.error("Error", err.responseJSON.message)
                } else {
                    toastr.error("Error", "ServerFail")
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
        data.port = parseInt(data.port)
        if (data.port < 0) {
            toastr.error('IP address does not exist', 'IP Error')
            return
        }
        $(".add-node-loading").removeClass("hide")
        $.ajax({
            type: "POST",
            url: "/nodes",
            data: JSON.stringify(data),
            success: function (response) {
                $(".add-node-loading").addClass("hide")
                // console.log(data)
                //     // Add to list 
                window.location.reload()
            },
            error: function (err) {
                $(".add-node-loading").addClass("hide")
                if (err && err.responseJSON && err.responseJSON.message) {
                    toastr.error("Add new node fail", err.responseJSON.message)
                } else {
                    toastr.error("Add new node fail", "ServerFail")
                }
            },
            contentType: "application/json"
        })
    })

    // 每隔2秒发送一次查询日志的请求
    setInterval(function () {
        $.ajax({
            type: "GET",
            url: "/status",
            success: function (response) {
                if (response && typeof response.data === "object") {
                    var result = response.data.filter(function (d) {
                        if (typeof d.result !== 'undefined' && d.result === true) {
                            console.log(d)
                            return true
                        }
                        return false
                    });
                    console.log(result.length)
                    if (result.length > 0) {
                        // 已经获得结果，关闭loading
                        $(".master-running").addClass("hide")
                    }
                    logger.batch(response.data)
                }
            },
            contentType: "application/json"
        })
    }, 2000)
})