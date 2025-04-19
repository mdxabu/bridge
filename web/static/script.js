window.onload = function () {
    // Fetch data every 2 seconds
    setInterval(fetchPingData, 2000);

    // Setup logout functionality
    document.getElementById('logout').onclick = function() {
        window.location.href = '/';
    };
};

function fetchPingData() {
    fetch('/api/data')
        .then(response => response.json())
        .then(data => {
            // Update charts with new data
            updateSentVsReceivedChart(data);
            updatePacketLossChart(data);
            updateRttChart(data);
        })
        .catch(error => console.error('Error fetching data:', error));
}

function updateSentVsReceivedChart(data) {
    const sentData = data.map(item => item.Sent);
    const receivedData = data.map(item => item.Received);

    Highcharts.chart('sent-vs-received', {
        chart: {
            type: 'line'
        },
        title: {
            text: 'Packets Sent & Received'
        },
        xAxis: {
            categories: data.map(item => new Date(item.Timestamp * 1000).toLocaleTimeString())
        },
        yAxis: {
            title: {
                text: 'Count'
            }
        },
        series: [{
            name: 'Sent',
            data: sentData
        }, {
            name: 'Received',
            data: receivedData
        }]
    });
}

function updatePacketLossChart(data) {
    const packetLossData = data.map(item => item.PacketLoss);

    Highcharts.chart('packet-loss', {
        chart: {
            type: 'line'
        },
        title: {
            text: 'Packet Loss'
        },
        xAxis: {
            categories: data.map(item => new Date(item.Timestamp * 1000).toLocaleTimeString())
        },
        yAxis: {
            title: {
                text: 'Loss (%)'
            }
        },
        series: [{
            name: 'Loss',
            data: packetLossData
        }]
    });
}

function updateRttChart(data) {
    const rttData = data.map(item => item.RTT);

    Highcharts.chart('rtt', {
        chart: {
            type: 'line'
        },
        title: {
            text: 'Round Trip Time (RTT)'
        },
        xAxis: {
            categories: data.map(item => new Date(item.Timestamp * 1000).toLocaleTimeString())
        },
        yAxis: {
            title: {
                text: 'RTT (ms)'
            }
        },
        series: [{
            name: 'RTT',
            data: rttData
        }]
    });
}
