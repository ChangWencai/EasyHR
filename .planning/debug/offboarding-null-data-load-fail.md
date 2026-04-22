---
name: offboarding-null-data-load-fail
status: investigating
trigger: "web离职管理界面，接口/api/v1/offboardings?page=1&page_size=2 返回的数据{\"code\": 0, \"data\": null, \"message\": \"success\", \"meta\": {\"page\": 1, \"page_size\": 20, \"total\": 0}}，其中data: null，web界面会出现\"加载失败\"提示"
created: 2026-04-22
updated: 2026-04-22
---

## Symptoms
- **Expected**: 接口成功返回 code=0，data 应为空数组 [] 或 null，界面应显示"暂无数据"或空状态
- **Actual**: data=null 时前端显示"加载失败"
- **Timeline**: 新功能，刚实现离职管理界面
- **Reproduction**: 访问离职管理页面时触发

## Current Focus
hypothesis: ""
next_action: "gather initial evidence"
