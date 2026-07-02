let selections = [];

const COMPARE_API_URL = '/api/v1/compare/';

function init() {
    initTheme();
    initCompareCheckboxes();
    isDisableChartButton(true);
    // тут же данные для инициализации - запросы таблиц и тд, если надо. 
}
function initTheme() {
    const isDarkTheme = !!localStorage.getItem('perf_dark_theme');
    main_theme_id.checked = isDarkTheme;
    changeTheme(isDarkTheme);
}

function changeTheme(event) {
    localStorage.setItem('perf_dark_theme', event? 'dark': '');
    if (event) {
        document.body.classList.add('dark');
    } else {
        document.body.classList.remove('dark');
    }
}
function toggleForm(isShow) {
    if (isShow) {
        form_get.style.display = '';
    } else {
        form_get.style.display = 'none';
    }
}
function toggleInfoBlock(isShow) {
    if (isShow) {
        info_block.style.display = '';
    } else {
        info_block.style.display = 'none';
    }
}
function isDisableChartButton(condition) { // вкл/выкл кнопку графиков
    show_chart_button.disabled = condition;
}
function getRowCompareId(row) {
    if (!row) return null;

    const checkbox = row.querySelector('input[name="compare"]');
    const compareCell = row.querySelector('td.check-multy[data-id]')
        || row.querySelector('td.check-multy')
        || row.querySelector('td.table-area-property');

    const rawId = compareCell?.getAttribute('data-id')
        || row.getAttribute('data-id')
        || checkbox?.value;

    if (!rawId) return null;

    const id = Number(rawId);
    return Number.isFinite(id) ? id : null;
}

function initCompareCheckboxes() {
    records_table.querySelectorAll('tbody > tr.table-area-object').forEach((row) => {
        const checkbox = row.querySelector('input[name="compare"]');
        const id = getRowCompareId(row);
        if (checkbox && id !== null) {
            checkbox.value = String(id);
        }
    });
}

function setSelections() {
    const list = records_table.querySelectorAll('tbody > tr.table-area-object');
    const checked = [];

    list.forEach((line) => {
        const checkbox = line.querySelector('input[name="compare"]');
        if (!checkbox || !checkbox.checked) return;

        const id = getRowCompareId(line);
        if (id !== null) {
            checked.push({ checkbox, id });
        }
    });

    while (checked.length > 2) {
        checked.shift().checkbox.checked = false;
    }

    selections = checked.map((item) => item.id);
    isDisableChartButton(selections.length !== 2);
}

function getBarChart( // возвращает SVG bar-chart
    data, // данные в формате [ [x_value, y_left_value, y_right_value],... ]
    xFormating = (d) => d, // функция форматирования значения перез отрисовкой на оси
    yFormating_l = (d) => d,
    yFormating_r = (d) => d,
    xAxisName = '', // label оси Х
    yAxisName_l = '',
    yAxisName_r = '',
    xScaleUnit = '', // Единица измерения на оси - конкатенируется с label
    yScaleUnit_l = '',
    yScaleUnit_r = '',
    header = '', // Заголовок графика
    description = [], // дополнительное текствое описание под графиком, данные в виде [ {title:'foo', value:'foo'},... ]
    xMarksLimit = 10, // примерное количество меток по оси X оставляемое после рендера - нужно для избежания каши значений под осью
    yColor_l = '#74d374', // цвет баров строящихся отнсительно шкалы y1
    yColor_r = '#61b4e0' // цвет баров строящихся отнсительно шкалы y2
    ) {
    const clipPathID = Math.random();
    const width = 928;
    const height = 500;
    const marginTop = 30;
    const marginRight = 70;
    const marginBottom = 120;
    const marginLeft = 70;
    
    // create scales
    const x = d3.scaleBand()
        .domain(d3.map(data, (d) => d[0]))
        .range([marginLeft, width - marginRight])
        .padding(0.55);

    const y = d3.scaleLinear()
        .domain([0, d3.max(data, (d) => +d[1]) * 1.2])
        .range([height - marginBottom, marginTop]);

    const y2 = d3.scaleLinear()
        .domain([0, d3.max(data, (d) => +d[2]) * 1.2])
        .range([height - marginBottom, marginTop]);

    // setup x axis view
    const xAxis = d3.axisBottom(x)
        .tickSizeOuter(0)
        .tickFormat(t => xFormating(t))
        .tickValues(x.domain().filter(function(d,i){ return !(i%getStepDivider(xMarksLimit))}));
  
    // Create the SVG container.
    const svg = d3.create('svg')
        .attr('id', 'svg_chart')
        .attr('class','chart')
        .attr('width', width)
        .attr('height', height)
        .attr('viewBox', [0, 0, width, height])
        .attr('style', 'max-width: 1000px; height: auto;')
        .style('pointer-events', 'none')
        .call(zoom);
        
    // adding a scrollable area
    svg.append('rect')
        .attr('class', 'focus')
        .attr('width', width - marginLeft - marginRight)
        .attr('height', height - marginTop - marginBottom)
        .attr('transform', `translate(${marginLeft}, ${marginTop})`)
        .style('fill', '#b2b2b2')
        .style('opacity', 0.2)
        .style('pointer-events', 'all');

    // trimming outgoing content
    svg.append('defs')
        .append('clipPath')
        .attr('id', `${clipPathID}`)
        .append('rect')
        .attr('width',  width - marginRight - marginLeft)
        .attr('height', height - marginTop)
        .attr('transform', `translate(${marginLeft},0)`);
  
    // Add a rect for each bar.
    svg.append('g')
        .attr('class', 'bars-l')
        .attr('fill', yColor_l)
        .style('clip-path', `url(#${clipPathID})`)
        .selectAll('rect')
            .data(data)
            .join('rect')
            .attr('class', 'bar')
            .attr('x', (d) => x(d[0]) - x.bandwidth()/2)
            .attr('y', (d) => y(d[1]))
            .attr('height', (d) => y(0) - y(d[1]))
            .attr('width', x.bandwidth())
            .on('mouseover', pointermoved)
            .on('mouseout', pointerleft);
    
    svg.append('g')
        .attr('class', 'bars-r')
        .attr('fill', yColor_r)
        .style('clip-path', `url(#${clipPathID})`)
        .selectAll('rect')
            .data(data)
            .join('rect')
            .attr('class', 'bar')
            .attr('x', (d) => x(d[0]) + x.bandwidth()/2)
            .attr('y', (d) => y2(d[2]))
            .attr('height', (d) => y2(0) - y2(d[2]))
            .attr('width', x.bandwidth())
            .on('mouseover', pointermoved)
            .on('mouseout', pointerleft);
  
    // Add the x-axis and label.
    svg.append('g')
        .attr('class', 'x-axis')
        .attr('transform', `translate(0,${height - marginBottom})`)
        .style('clip-path', `url(#${clipPathID})`)
        .call(xAxis)
        .call(g => g.append('text')
            .attr('x', width-marginRight-marginLeft)
            .attr('y', 40)
            .attr('fill', 'currentColor')
            .attr('text-anchor', 'start')
            .text(`→ ${xAxisName} ${xScaleUnit}`)
        );
  
    // Add the y-axis and label
    svg.append('g')
        .attr('class', 'y-axis')
        .attr('transform', `translate(${marginLeft},0)`)
        .call(d3.axisLeft(y).tickFormat(t => yFormating_l(t)))
        .call(g => g.append('text')
            .attr('x', -marginLeft)
            .attr('y', 10)
            .attr('fill', 'currentColor')
            .attr('text-anchor', 'start')
            .text(`↑ ${yAxisName_l} ${yScaleUnit_l}`)
        );

    // Add the second y-axis and label
    svg.append('g')
        .attr('class', 'y-axis')
        .attr('transform', `translate(${width - marginRight},0)`)
        .call(d3.axisRight(y2).tickFormat(t => yFormating_r(t)))
        .call(g => g.append('text')
            .attr('x', -30)
            .attr('y', 10)
            .attr('fill', 'currentColor')
            .attr('text-anchor', 'start')
            .text(`↑ ${yAxisName_r} ${yScaleUnit_r}`)
        );

    // add details
    svg.append('text')
        .attr('class', 'chartHeader')
        .attr('x', '50%')
        .attr('y', 10)
        .attr('fill', 'currentColor')
        .attr('text-anchor', 'middle')
        .text(`${header}`);

    const details = svg.append('g')
        .attr('class', 'details')
        .attr('transform', `translate(0,${height - marginBottom + 70})`)
        .selectAll('text')
            .data([xAxisName, yAxisName_l, yAxisName_r])
            .join('text')
            .attr('x', marginLeft)
            .attr('y', (d, i) => i * 20)
            .attr('fill', 'currentColor')
            .attr('text-anchor', 'start')
            .style('transition', 'all .2s')
            .style('opacity', 0.7)
            .text((t) => t + ':    ...');
        
    svg.append('g')
        .attr('class', 'description')
        .attr('transform', `translate(0,${height - marginBottom + 70})`)
        .selectAll('text')
            .data(description)
            .join('text')
            .attr('x', marginLeft + 150)
            .attr('y', (d, i) => i * 20)
            .attr('fill', 'currentColor')
            .attr('text-anchor', 'start')
            .text((t) => t.title + t.value);

    // testing
    // svg.selectAll('.bars-l')
    //     .selectAll('text')
    //         .data(data)
    //         .join('text')
    //         .attr('class', 'bar-label')
    //         .attr('x', (d) => x(d[0]))
    //         .attr('y', (d) => y(d[1]) - 5)
    //         .attr('textLength', x.bandwidth())
    //         // .attr('x', (d) => x(d[0])+x.bandwidth()/2)
    //         // .attr('y', (d) => (y(d[1])<y2(d[2])?y(d[1]):y2(d[2])) - 5)
    //         // .attr('textLength', (x.bandwidth() * 2) > 90? 90: x.bandwidth() * 2)
    //         .attr('lengthAdjust', 'spacingAndGlyphs')
    //         .attr('text-anchor', 'middle')
    //         .attr('startOffset', '50%')
    //         .style('font-size', '10px')
    //         // .text(d => d[0]);
    //         .text(d => yFormating_l(Math.round(d[1])));
    
    // svg.selectAll('.bars-r')
    //     .selectAll('text')
    //         .data(data)
    //         .join('text')
    //         .attr('class', 'bar-label')
    //         .attr('x', (d) => x(d[0]) + x.bandwidth())
    //         .attr('y', (d) => y2(d[2]) - 5)
    //         .attr('textLength', x.bandwidth())
    //         .attr('lengthAdjust', 'spacingAndGlyphs')
    //         .attr('text-anchor', 'middle')
    //         .attr('startOffset', '50%')
    //         .style('font-size', '10px')
    //         .text(d => yFormating_l(d[2]));

    // Return the SVG element.
    return svg.node();

    // Sweep values ​​along the x-axis 
    function getStepDivider(maxDisplayCount) {
        const domainLength = x.domain().length;
        const viewportWidth = width-marginLeft-marginRight;
        const zoomedViewport = x.range()[1] - x.range()[0];
        const viewportDifference = zoomedViewport / viewportWidth;
        const k = Math.round(domainLength / (maxDisplayCount * viewportDifference));

        return k;
    }

    // tooltip
    function pointermoved(event, data) {
        details.style('opacity', 1).text((t,i) => t + `: ${data[i]}`);
    }
    function pointerleft() {
        details.style('opacity', 0.7).text(t => t + ': ...');
    }

    // zoom events
    function zoom(svg) {
        const extent = [[marginLeft, marginTop], [width - marginLeft, height - marginTop]];
    
        svg.call(d3.zoom()
            .scaleExtent([1, 8])
            .translateExtent(extent)
            .extent(extent)
            .on('zoom', zoomed));
    
        function zoomed(event) {
            xAxis.tickValues(x.domain().filter(function(d,i){ return !(i%getStepDivider(xMarksLimit))}));

            x.range([marginLeft, width - marginLeft].map(d => event.transform.applyX(d)));
            svg.selectAll('.bars-l rect').attr('x', (d,i) => x(d[0]) - x.bandwidth()/2).attr('width', x.bandwidth());
            svg.selectAll('.bars-r rect').attr('x', (d,i) => x(d[0]) + x.bandwidth()/2).attr('width', x.bandwidth());
            svg.selectAll('.x-axis').call(xAxis);

            // testing
            // svg.selectAll('.bars-l text').attr('x', (d,i) => x(d[0])+x.bandwidth()/2).attr('textLength', (x.bandwidth() * 2) > 90? 90: x.bandwidth() * 2);
            // svg.selectAll('.bars-l text').attr('x', (d,i) => x(d[0])).attr('textLength', x.bandwidth());
            // svg.selectAll('.bars-r text').attr('x', (d,i) => x(d[0])+x.bandwidth()).attr('textLength', x.bandwidth());
        }
    }
}
function showPerfomanceChart(data) {
    if (!data) return;
    if (!data[0]) return;

    const hardware = data[0]['hardware']? data[0]['hardware']: '';
    const hardwaremain = hardware.split(';').shift(); // выделение основной подстроки - прим. Intel(R)/Intel(R) Xeon(R) CPU E5-2670 0 @ 2.60GHz CPUs:32
    const testProfile = data[0]['test_profile']? data[0]['test_profile']: '';

    const metric = data[0]['fio_data'];
    const xValueFormat = (t) => (t/1000).toFixed(1); // приведение к секундам
    // приведение к виду 1К..К 1М..М
    const yValueFormat = (t) => {
        let litterK = '';
        let litterM = '';
        while (t/1000 >= 1) {
            litterK += 'K';
            t = t/1000;
        }

        if (litterK.length%2 == 0 && litterK.length) {
            let count = litterK.length/2;
            while (count) {
                count--;
                litterM += 'M';
            }
        }

        if (litterM.length) {
            t += litterM;
        } else {
            t += litterK;
        }

        return t;
    };
    // полуение svg и добавлеие в container
    const node = getBarChart(
        metric,
        xValueFormat,
        yValueFormat,
        yValueFormat,
        'Time',
        'Bandwidth',
        'IOPS',
        '(s)',
        '(byte)',
        '',
        data[0]['annotation']? data[0]['annotation']: '',
        [
            {title: 'Hardware: ', value: hardwaremain},
            {title: '', value: hardware.replace(hardwaremain + ';', '')},
            {title:'Test profile: ', value: testProfile}
        ]
    );
    container.append(node);
}
function parceCompareData(data) { // парсинг данных перед передачей в функцию рисования
    const metric = [];
    const xDomain = [];
    
    data.forEach((d, i) => {
        metric.push(
            [
                !xDomain.find(domain => domain == d['annotation'])?
                    d['annotation']:
                    `${d['annotation']} (${xDomain.filter(domain => domain == d['annotation']).length})`,
                d['fio_data'].reduce((acc, fio) => acc += +fio[1],0) / d['fio_data'].length,
                d['fio_data'].reduce((acc, fio) => acc += +fio[2],0) / d['fio_data'].length
            ]
        );
        xDomain.push(d['annotation']);
    });

    return metric;
}
function showCompareChart(data) {
    if (!data) return;
    if (data.length < 2) return;

    const xValueFormat = (d) => d;
    const yValueFormat = (t) => {
        let litterK = '';
        let litterM = '';
        while (t/1000 >= 1) {
            litterK += 'K';
            t = t/1000;
        }

        if (litterK.length%2 == 0 && litterK.length) {
            let count = litterK.length/2;
            while (count) {
                count--;
                litterM += 'M';
            }
        }

        if (litterM.length) {
            t += litterM;
        } else {
            t += litterK;
        }

        return t;
    };
    const metric = parceCompareData(data);

    const node = getBarChart(
        metric,
        xValueFormat,
        yValueFormat,
        yValueFormat,
        'Оборудование',
        'Чтение',
        'Запись',
        '',
        '(Кб/с)',
        '(Кб/s)',
        'Сравнить',
        [],
        4,
        '#6767c6',
        '#d36666'
    );
    container.append(node);
}

async function getChartData(url, options) { // запрос данных от бэка
    try {
        const response = await fetch(url, options);
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        return await response.json();
    } catch (err) {
        console.error('Compare request failed:', err);
        return null;
    }
}

async function onShowChartButton() { // метод по кнопке
    setSelections();

    if (selections.length !== 2) return;

    const payload = { ids: selections };
    const options = {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    };

    isDisableChartButton(true);

    const data = await getChartData(COMPARE_API_URL, options);

    if (data && data.status === 'ok' && data.url) {
        window.open(data.url, '_blank');
    } else {
        alert('Не удалось получить URL для сравнения. Проверьте, что nginx проксирует /api/v1/compare/ на rest-graph-api.');
    }

    isDisableChartButton(selections.length !== 2);
}

init();
