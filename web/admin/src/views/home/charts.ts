/**
 * 工作台图表 option：登录趋势（成功 / 失败 / 锁定）。
 */
import type { EChartsOption } from 'echarts'

import { emptyLoginTrend, type LoginTrend } from './mock'

export function buildLoginTrendOption(trend?: LoginTrend): EChartsOption {
  const data = trend ?? emptyLoginTrend()
  return {
    color: ['#1677ff', '#ff4d4f', '#faad14'],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross', label: { backgroundColor: '#6a7985' } },
    },
    legend: {
      data: ['成功', '失败', '锁定'],
      right: 0,
      top: 0,
      icon: 'circle',
      itemWidth: 8,
      itemHeight: 8,
    },
    grid: { left: 40, right: 16, top: 36, bottom: 28 },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data.dates,
      axisTick: { show: false },
      axisLine: { lineStyle: { color: '#f0f0f0' } },
      axisLabel: { color: 'rgba(0,0,0,0.45)' },
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      splitLine: { lineStyle: { type: 'dashed', color: '#f0f0f0' } },
      axisLabel: { color: 'rgba(0,0,0,0.45)' },
    },
    series: [
      {
        name: '成功',
        type: 'line',
        smooth: true,
        showSymbol: false,
        areaStyle: { opacity: 0.1 },
        data: data.success,
      },
      {
        name: '失败',
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: data.fail,
      },
      {
        name: '锁定',
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: data.locked,
      },
    ],
  }
}
