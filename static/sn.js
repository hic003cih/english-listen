var form = document.getElementById('snForm');
form.onsubmit = function(e) {
    e.preventDefault();
    var data = document.getElementById('sninput').value;
    var div = document.getElementById('snresultDiv');
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/sntransfer', true);
    xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
    xhr.onreadystatechange = function() {
        if(this.readyState !== XMLHttpRequest.DONE) {
            return;
        }
        var response = JSON.parse(this.responseText);
        if(response.error) {
            div.innerHTML = '<p>' + response.error + '</p>';
        } else {
            div.innerHTML = '<pre>' + response.result + '</pre>';
        }
    };
    xhr.send('data=' + encodeURIComponent(data));
};