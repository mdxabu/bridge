let metricsData = [];
let tooltipEl;
let totalPacketsSent = 0;
let totalPacketsReceived = 0;
let showCompleteTimeline = true;

const COLORS = {
    SENT: 'rgb(87, 148, 242)',
    RECEIVED: 'rgb(92, 184, 92)',
    LOSS: 'rgb(217, 83, 79)',
    RTT: 'rgb(240, 173, 78)',
    GRID: 'rgba(255, 255, 255, 0.1)',
    TEXT: 'rgba(255, 255, 255, 0.5)',
    BACKGROUND: 'rgba(0, 0, 0, 0.7)'
};

function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

window.onload = function() {
    tooltipEl = document.createElement('div');
    tooltipEl.className = 'tooltip';
    document.body.appendChild(tooltipEl);
    
    fetchMetricsData();
    
    setInterval(fetchMetricsData, 5000);
    
    window.addEventListener('resize', updateDashboard);
    
    document.getElementById('logout').addEventListener('click', function(e) {
        e.preventDefault();
        window.location.href = '/logout';
    });
    
    const tableContainer = document.querySelector('.table-container');
    if (tableContainer) {
        tableContainer.style.maxHeight = 'none';
        tableContainer.style.overflow = 'visible';
    }

    startMetricsCollection();
    
    document.getElementById('export-btn').addEventListener('click', function() {
        const a = document.createElement('a');
        a.href = '/api/data';
        a.download = 'bridge-metrics-' + new Date().toISOString().split('T')[0] + '.json';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
    });
};

function createHoverElements() {
    const hoverLine = document.createElement('div');
    hoverLine.className = 'hover-line';
    hoverLine.style.display = 'none';
    document.body.appendChild(hoverLine);
    
    const packetsSent = document.createElement('div');
    packetsSent.className = 'data-point sent-point';
    packetsSent.style.display = 'none';
    document.body.appendChild(packetsSent);
    
    const packetsReceived = document.createElement('div');
    packetsReceived.className = 'data-point received-point';
    packetsReceived.style.display = 'none';
    document.body.appendChild(packetsReceived);
    
    const lossPoint = document.createElement('div');
    lossPoint.className = 'data-point loss-point';
    lossPoint.style.display = 'none';
    document.body.appendChild(lossPoint);
    
    const rttPoint = document.createElement('div');
    rttPoint.className = 'data-point rtt-point';
    rttPoint.style.display = 'none';
    document.body.appendChild(rttPoint);
}

function fetchMetricsData() {
    fetch('/api/data')
        .then(response => response.json())
        .then(data => {
            if (Array.isArray(data)) {
                metricsData = data;
                updateDashboard();
            } else {
                console.error('Invalid data format:', data);
            }
        })
        .catch(error => {
            console.error('Error fetching metrics data:', error);
        });
}

function updateDashboard() {
    if (metricsData.length === 0) return;
    
    updateStats();
    updateSummaryStats();
    
    drawPacketsGraph('packets-graph', showCompleteTimeline);
    drawLossGraph('loss-graph', showCompleteTimeline);
    drawRttGraph('rtt-graph', showCompleteTimeline);
}

function updateStats() {
    if (metricsData.length === 0) return;
    
    const currentSent = metricsData.reduce((sum, data) => sum + data.sent, 0);
    const currentReceived = metricsData.reduce((sum, data) => sum + data.received, 0);
    
    if (currentSent > totalPacketsSent) totalPacketsSent = currentSent;
    if (currentReceived > totalPacketsReceived) totalPacketsReceived = currentReceived;
    
    document.getElementById('sent-value').textContent = totalPacketsSent;
    document.getElementById('received-value').textContent = totalPacketsReceived;
    
    const latest = metricsData[metricsData.length - 1];
    
    const lossEl = document.getElementById('loss-value');
    const loss = latest.loss !== undefined ? latest.loss : 
        (latest.sent > 0 ? ((latest.sent - latest.received) / latest.sent) * 100 : 0);
    
    lossEl.textContent = `${loss.toFixed(1)}%`;
    
    if (loss > 10) {
        lossEl.className = 'value stat-danger';
    } else if (loss > 0) {
        lossEl.className = 'value stat-warning';
    } else {
        lossEl.className = 'value stat-success';
    }
    
    const rttEl = document.getElementById('rtt-value');
    const rttValue = latest.rtt;
    rttEl.textContent = `${rttValue} ms`;
    
    updateTrafficTable();
}

function updateSummaryStats() {
    if (metricsData.length === 0) return;
    
    const totalRtt = metricsData.reduce((sum, data) => sum + (data.rtt_ms !== undefined ? data.rtt_ms : data.rtt), 0);
    const avgRtt = totalRtt / metricsData.length;
    document.getElementById('avg-rtt').textContent = `${avgRtt.toFixed(2)} ms`;
    
    const totalSent = metricsData.reduce((sum, data) => sum + data.sent, 0);
    const totalReceived = metricsData.reduce((sum, data) => sum + data.received, 0);
    
    const totalLoss = totalSent > 0 ? ((totalSent - totalReceived) / totalSent) * 100 : 0;
    const successRate = totalSent > 0 ? (totalReceived / totalSent) * 100 : 0;
    
    document.getElementById('total-loss').textContent = `${totalLoss.toFixed(1)}%`;
    document.getElementById('success-rate').textContent = `${successRate.toFixed(1)}%`;
    
    const totalLossEl = document.getElementById('total-loss');
    if (totalLoss > 10) {
        totalLossEl.className = 'summary-value stat-danger';
    } else if (totalLoss > 0) {
        totalLossEl.className = 'summary-value stat-warning';
    } else {
        totalLossEl.className = 'summary-value stat-success';
    }
    
    const successRateEl = document.getElementById('success-rate');
    if (successRate > 95) {
        successRateEl.className = 'summary-value stat-success';
    } else if (successRate > 85) {
        successRateEl.className = 'summary-value stat-warning';
    } else {
        successRateEl.className = 'summary-value stat-danger';
    }
}

function updateTrafficTable() {
    const tableBody = document.getElementById('traffic-data');
    if (!tableBody) return;
    
    tableBody.innerHTML = '';
    
    const displayData = metricsData;
    
    displayData.forEach(data => {
        const row = document.createElement('tr');
        
        row.innerHTML = `
            <td>${data.source || '127.0.0.1'}</td>
            <td>${data.destination || '8.8.8.8'}</td>
            <td>${data.sent}</td>
            <td>${data.received}</td>
            <td>${data.rtt} ms</td>
            <td>${formatTime(data.timestamp)}</td>
        `;
        
        tableBody.appendChild(row);
    });
    
    const tableContainer = document.querySelector('.table-container');
    if (tableContainer) {
        tableContainer.style.maxHeight = 'none';
        tableContainer.style.overflow = 'visible';
    }
}

function drawPacketsGraph(canvasId, showFullTimeline = true) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const {width, height} = setupCanvas(canvas);
    
    ctx.clearRect(0, 0, width, height);
    
    const drawWidth = width - 40;
    const drawHeight = height - 30;
    
    ctx.translate(30, 10);
    
    const dataToUse = showFullTimeline ? metricsData : metricsData;
    
    const xScale = drawWidth / (dataToUse.length - 1 || 1);
    const maxY = Math.max(
        ...dataToUse.map(d => Math.max(d.sent, d.received))
    ) * 1.1;
    const yScale = drawHeight / (maxY || 1);
    
    drawGrid(ctx, drawWidth, drawHeight, maxY, dataToUse);
    
    ctx.strokeStyle = COLORS.SENT;
    ctx.lineWidth = 2;
    drawLine(ctx, dataToUse, d => d.sent, xScale, yScale, drawHeight);
    
    ctx.strokeStyle = COLORS.RECEIVED;
    drawLine(ctx, dataToUse, d => d.received, xScale, yScale, drawHeight);
    
    setupHover(canvas, dataToUse, (data) => {
        return `Time: ${formatTime(data.timestamp)}<br>` +
               `Sent: ${data.sent}<br>` +
               `Received: ${data.received}`;
    });
    
    ctx.setTransform(1, 0, 0, 1, 0, 0);
}

function setupCanvas(canvas) {
    const container = canvas.parentElement;
    const devicePixelRatio = window.devicePixelRatio || 1;
    
    const rect = container.getBoundingClientRect();
    
    canvas.style.width = `${rect.width}px`;
    canvas.style.height = `${rect.height}px`;
    
    canvas.width = rect.width * devicePixelRatio;
    canvas.height = rect.height * devicePixelRatio;
    
    const ctx = canvas.getContext('2d');
    ctx.scale(devicePixelRatio, devicePixelRatio);
    
    return {width: rect.width, height: rect.height};
}

function drawLossGraph(canvasId, showFullTimeline = true) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const {width, height} = setupCanvas(canvas);
    
    ctx.clearRect(0, 0, width, height);
    
    const drawWidth = width - 40;
    const drawHeight = height - 30;
    
    ctx.translate(30, 10);
    
    const dataToUse = showFullTimeline ? metricsData : metricsData;
    
    const maxLoss = 100;
    const yScale = drawHeight / maxLoss;
    
    drawGrid(ctx, drawWidth, drawHeight, maxLoss, dataToUse);
    
    const xScale = drawWidth / (dataToUse.length - 1 || 1);
    
    ctx.fillStyle = 'rgba(217, 83, 79, 0.2)';
    ctx.beginPath();
    ctx.moveTo(0, drawHeight);
    
    dataToUse.forEach((data, i) => {
        const x = i * xScale;
        const lossValue = data.packet_loss !== undefined ? data.packet_loss : data.loss;
        const y = drawHeight - (lossValue * yScale);
        ctx.lineTo(x, y);
    });
    
    ctx.lineTo((dataToUse.length - 1) * xScale, drawHeight);
    ctx.closePath();
    ctx.fill();
    
    ctx.strokeStyle = COLORS.LOSS;
    ctx.lineWidth = 2;
    drawLine(ctx, dataToUse, d => d.packet_loss !== undefined ? d.packet_loss : d.loss, xScale, yScale, drawHeight);
    
    setupHover(canvas, dataToUse, (data) => {
        const lossValue = data.packet_loss !== undefined ? data.packet_loss : data.loss;
        return `Time: ${formatTime(data.timestamp)}<br>` +
               `Loss: ${lossValue.toFixed(1)}%`;
    });
    
    ctx.setTransform(1, 0, 0, 1, 0, 0);
}

function drawRttGraph(canvasId, showFullTimeline = true) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const {width, height} = setupCanvas(canvas);
    
    ctx.clearRect(0, 0, width, height);
    
    const drawWidth = width - 40;
    const drawHeight = height - 30;
    
    ctx.translate(30, 10);
    
    const dataToUse = showFullTimeline ? metricsData : metricsData;
    
    const xScale = drawWidth / (dataToUse.length - 1 || 1);
    const maxRtt = Math.max(...dataToUse.map(d => {
        return d.rtt_ms !== undefined ? d.rtt_ms : d.rtt;
    })) * 1.1;
    const yScale = drawHeight / (maxRtt || 1);
    
    drawGrid(ctx, drawWidth, drawHeight, maxRtt, dataToUse);
    
    ctx.strokeStyle = COLORS.RTT;
    ctx.lineWidth = 2;
    drawLine(ctx, dataToUse, d => d.rtt_ms !== undefined ? d.rtt_ms : d.rtt, xScale, yScale, drawHeight);
    
    setupHover(canvas, dataToUse, (data) => {
        const rttValue = data.rtt_ms !== undefined ? data.rtt_ms : data.rtt;
        return `Time: ${formatTime(data.timestamp)}<br>` +
               `RTT: ${rttValue.toFixed(2)} ms`;
    });
    
    ctx.setTransform(1, 0, 0, 1, 0, 0);
}

function drawGrid(ctx, width, height, maxValue, data) {
    ctx.strokeStyle = COLORS.GRID;
    ctx.lineWidth = 1;
    
    const yStep = height / 5;
    for (let y = 0; y <= height; y += yStep) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
    }
    
    const numVerticalLines = Math.min(data.length, 10);
    const xStep = width / numVerticalLines;
    
    for (let i = 0; i <= numVerticalLines; i++) {
        const x = i * xStep;
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
        ctx.stroke();
    }
    
    ctx.fillStyle = COLORS.TEXT;
    ctx.font = '12px Roboto';
    ctx.textAlign = 'right';
    ctx.textBaseline = 'middle';
    
    for (let i = 0; i <= 5; i++) {
        const y = height - (i * yStep);
        let label = '';
        
        if (maxValue !== undefined) {
            const value = (i * maxValue / 5);
            if (ctx.canvas.id === 'loss-graph') {
                label = value.toFixed(0) + '%';
            } else if (ctx.canvas.id === 'rtt-graph') {
                label = value.toFixed(1) + ' ms';
            } else {
                label = value.toFixed(0);
            }
        }
        
        const textWidth = ctx.measureText(label).width;
        ctx.fillStyle = COLORS.BACKGROUND;
        ctx.fillRect(-textWidth - 8, y - 10, textWidth + 6, 20);
        
        ctx.fillStyle = 'rgba(255, 255, 255, 0.9)';
        ctx.fillText(label, -5, y);
    }
    
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    
    for (let i = 0; i <= numVerticalLines; i++) {
        const x = i * xStep;
        const dataIndex = Math.floor(i * (data.length - 1) / numVerticalLines);
        
        if (dataIndex < data.length) {
            const timeLabel = formatTimeShort(data[dataIndex].timestamp);
            
            const textWidth = ctx.measureText(timeLabel).width;
            ctx.fillStyle = COLORS.BACKGROUND;
            ctx.fillRect(x - textWidth/2 - 2, height + 2, textWidth + 4, 16);
            
            ctx.fillStyle = 'rgba(255, 255, 255, 0.9)';
            ctx.fillText(timeLabel, x, height + 5);
        }
    }
    
    ctx.fillStyle = COLORS.TEXT;
    ctx.font = '12px Roboto';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    
    ctx.save();
    ctx.translate(-20, height / 2);
    ctx.rotate(-Math.PI / 2);
    
    if (ctx.canvas.id === 'packets-graph') {
        ctx.fillText('Packets', 0, 0);
    } else if (ctx.canvas.id === 'loss-graph') {
        ctx.fillText('Loss %', 0, 0);
    } else if (ctx.canvas.id === 'rtt-graph') {
        ctx.fillText('RTT (ms)', 0, 0);
    }
    
    ctx.restore();
    
    ctx.fillText('Time', width / 2, height + 20);
}

function drawLine(ctx, data, valueFunc, xScale, yScale, height) {
    const skipFactor = data.length > 1000 ? Math.floor(data.length / 1000) : 1;
    
    ctx.beginPath();
    let firstPoint = true;
    
    for (let i = 0; i < data.length; i += skipFactor) {
        const d = data[i];
        const x = (i / (data.length - 1)) * (data.length - 1) * xScale;
        const y = height - (valueFunc(d) * yScale);
        
        if (firstPoint) {
            ctx.moveTo(x, y);
            firstPoint = false;
        } else {
            ctx.lineTo(x, y);
        }
    }
    
    if (data.length > 0 && (data.length - 1) % skipFactor !== 0) {
        const lastData = data[data.length - 1];
        const x = (data.length - 1) * xScale;
        const y = height - (valueFunc(lastData) * yScale);
        ctx.lineTo(x, y);
    }
    
    ctx.stroke();
}

function setupHover(canvas, data, tooltipFormatter) {
    const container = canvas.parentElement;
    
    const hoverLine = document.createElement('div');
    hoverLine.className = 'hover-line';
    hoverLine.style.display = 'none';
    container.appendChild(hoverLine);
    
    let dataPoints = [];
    
    if (canvas.id === 'packets-graph') {
        dataPoints.push(createDataPoint(container, 'sent-point'));
        dataPoints.push(createDataPoint(container, 'received-point'));
    } else if (canvas.id === 'loss-graph') {
        dataPoints.push(createDataPoint(container, 'loss-point'));
    } else if (canvas.id === 'rtt-graph') {
        dataPoints.push(createDataPoint(container, 'rtt-point'));
    }
    
    canvas.addEventListener('mousemove', throttle(function(e) {
        const rect = canvas.getBoundingClientRect();
        
        const mouseX = e.clientX - rect.left;
        const mouseY = e.clientY - rect.top;
        
        if (mouseX < 0 || mouseX > rect.width || mouseY < 0 || mouseY > rect.height) {
            return;
        }
        
        const relativeX = mouseX / rect.width;
        const index = Math.min(
            Math.max(0, Math.round(relativeX * (data.length - 1))),
            data.length - 1
        );
        
        const dataPoint = data[index];
        if (!dataPoint) return;
        
        hoverLine.style.display = 'block';
        hoverLine.style.left = `${mouseX}px`;
        
        const tooltipContent = tooltipFormatter(dataPoint);
        tooltipEl.innerHTML = tooltipContent;
        tooltipEl.style.display = 'block';
        tooltipEl.style.left = `${e.clientX + 15}px`;
        tooltipEl.style.top = `${e.clientY}px`;
        
        if (canvas.id === 'packets-graph') {
            const maxY = Math.max(...data.map(d => Math.max(d.sent, d.received))) * 1.1;
            const drawableHeight = rect.height - 40;
            
            const sentY = drawableHeight - ((dataPoint.sent / maxY) * drawableHeight) + 10;
            dataPoints[0].style.display = 'block';
            dataPoints[0].style.left = `${mouseX}px`;
            dataPoints[0].style.top = `${sentY}px`;
            
            const receivedY = drawableHeight - ((dataPoint.received / maxY) * drawableHeight) + 10;
            dataPoints[1].style.display = 'block';
            dataPoints[1].style.left = `${mouseX}px`;
            dataPoints[1].style.top = `${receivedY}px`;
        } else if (canvas.id === 'loss-graph') {
            const drawableHeight = rect.height - 40;
            const y = drawableHeight - ((dataPoint.loss / 100) * drawableHeight) + 10;
            
            dataPoints[0].style.display = 'block';
            dataPoints[0].style.left = `${mouseX}px`;
            dataPoints[0].style.top = `${y}px`;
        } else if (canvas.id === 'rtt-graph') {
            const maxRtt = Math.max(...data.map(d => d.rtt)) * 1.1;
            const drawableHeight = rect.height - 40;
            const y = drawableHeight - ((dataPoint.rtt / maxRtt) * drawableHeight) + 10;
            
            dataPoints[0].style.display = 'block';
            dataPoints[0].style.left = `${mouseX}px`;
            dataPoints[0].style.top = `${y}px`;
        }
    }, 16));
    
    canvas.addEventListener('mouseout', function() {
        tooltipEl.style.display = 'none';
        hoverLine.style.display = 'none';
        
        dataPoints.forEach(point => {
            point.style.display = 'none';
        });
    });
    
    function createDataPoint(container, className) {
        const point = document.createElement('div');
        point.className = `data-point ${className}`;
        point.style.display = 'none';
        container.appendChild(point);
        return point;
    }
}

function startTranslation() {
    fetch('/api/start-translation', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to start translation');
        }
        return response.json();
    })
    .then(data => {
        console.log('NAT64 translation started successfully:', data);
    })
    .catch(error => {
        console.error('Error starting NAT64 translation:', error);
    });
}

function startMetricsCollection() {
    fetch('/api/start-metrics')
        .then(response => response.json())
        .then(data => {
            console.log('Metrics collection started:', data);
        })
        .catch(error => {
            console.error('Failed to start metrics collection:', error);
        });
}

function formatTime(timestamp) {
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString();
}

function formatTimeShort(timestamp) {
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
