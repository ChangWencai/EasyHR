---
status: fixing
trigger: 组织架构只有部门，没有岗位，且在更新部门名称时，编辑框在最左上角，位置不对
symptoms:
  expected_behavior: "组织架构树形结构应同时显示部门和岗位节点；更新部门名称时编辑框应出现在部门名称旁边"
  actual_behavior: "组织架构只显示部门，不显示岗位；编辑部门名称时输入框出现在页面最左上角"
  error_messages: "无明显错误信息"
  timeline: "新功能或近期改动"
  reproduction: "进入组织架构管理页面，查看树形结构；点击编辑部门名称"
created: "2026-04-22"
updated: "2026-04-22T17:12:00+08:00"
---

## Current Focus

hypothesis: "根因3（已有数据不显示岗位）：MigrateFromEmployeePositions 仅创建了 Position 表记录，未反向更新 employee.position_id，导致 BuildTree 按 position_id 分组时所有历史员工的 PositionID 均为 NULL，岗位节点无员工子节点因此在树中不可见。"
test: "检查 GetTree 中 MigrateFromEmployeePositions 调用后是否更新了 employee.position_id"
expecting: "GetTree 中新增 backfill 逻辑后，已有员工的 position_id 被正确关联，岗位节点应显示在组织架构树中"
next_action: "已在 GetTree 中添加历史数据 backfill 逻辑（position_id 为 NULL 的员工按 position 文本关联到对应岗位），go build 通过，等待用户验证"

## Symptoms
<!-- IMMUTABLE -->
expected_behavior: "组织架构树形结构应同时显示部门和岗位节点；更新部门名称时编辑框应出现在部门名称旁边"
actual_behavior: "组织架构只显示部门，不显示岗位；编辑部门名称时输入框出现在页面最左上角"
error_messages: "无明显错误信息"
timeline: "新功能或近期改动"
reproduction: "进入组织架构管理页面，查看树形结构；点击编辑部门名称"

## Eliminated
<!-- APPEND only -->
- hypothesis: "context-menu 模板中缺少删除部门项"
  evidence: "模板第85-87行完整包含删除部门项（@click=showDeleteTransferDialog），并非模板缺失"
  timestamp: "2026-04-22T17:00:00"
- hypothesis: "ECharts offsetX/offsetY 坐标系与 CSS 不匹配导致位置错误（旧根因1，已部分修复）"
  evidence: "已改用 position:fixed + getBoundingClientRect() 转换坐标，理论上应解决位置错误问题"
  timestamp: "2026-04-22T17:02:00"
- hypothesis: "前端 CSS position 从 absolute 改为 fixed 导致输入框消失"
  evidence: "position:fixed 是正确的修复（结合 getBoundingClientRect），不是导致问题的原因"
  timestamp: "2026-04-22T17:03:00"

## Evidence
<!-- APPEND only -->
- timestamp: "2026-04-22T17:00:00"
  checked: "OrgChart.vue template (lines 75-104)"
  found: "删除部门项完整存在于 template (line 85-87: @click=showDeleteTransferDialog, el-icon Delete)。上下文菜单由 v-if=contextMenuVisible 控制，数据绑定到 contextMenuX/Y（来源于 offsetX + getBoundingClientRect()）"
  implication: "模板无误，问题在 CSS（可能被遮挡）或数据绑定（contextMenuVisible 可能为 false）"
- timestamp: "2026-04-22T17:02:00"
  checked: "OrgChart.vue CSS (lines 617-628)"
  found: ".context-menu { position: fixed; z-index: 9999; } 和 .inline-edit-overlay { position: fixed; z-index: 9998; } 样式均已正确应用"
  implication: "CSS 样式正确，fixed 定位已设置。问题可能是 ECharts offsetX/offsetY 值为 0（click 事件未正确传递坐标），或 chart 容器的 getBoundingClientRect() 返回 (0,0)"
- timestamp: "2026-04-22T17:03:00"
  checked: "git diff frontend/src/views/employee/OrgChart.vue"
  found: "CSS position 改为 fixed 已应用；getChartViewportOffset() 坐标转换已添加；坐标绑定从直接 offsetX/offsetY 改为 (offsetX??0)+offset.left"
  implication: "所有修复均已正确应用。但 offsetX/offsetY 在 Vue ECharts 中可能为 canvas 左上角坐标而非节点实际坐标"
- timestamp: "2026-04-22T17:04:00"
  checked: "Go backend diff"
  found: "所有 backend 修复已正确应用：dto.go PositionID 字段、service.go PositionID 关联逻辑、position service FindOrCreateByName、main.go 模块顺序调整。go build 通过"
  implication: "Backend 代码正确，但 npm run build 前端编译报错（TS6133/TS2322 等错误，与 OrgChart.vue 无关）"
- timestamp: "2026-04-22T17:10:00"
  checked: "OrgChart.vue bindChartEvents() - event.clientX vs offsetX"
  found: "ECharts click/contextmenu 事件的 params.event 是原生 MouseEvent，包含 clientX/clientY（viewport 坐标）。event.clientX 相比 offsetX 更可靠，因为 offsetX 是 canvas 内部坐标且在 vue-echarts 中可能不可用或为 0。已将 event.clientX/clientY 作为主坐标源（fallback 到 offsetX + getBoundingClientRect）"
  implication: "event.clientX 直接给出鼠标在视口中的坐标，position:fixed 元素使用该坐标后，编辑框和右键菜单应出现在鼠标点击处，而非 (0,0)"
- timestamp: "2026-04-22T18:20:00"
  checked: "department/service.go GetTree() - MigrateFromEmployeePositions 与 BuildTree 的关联"
  found: "MigrateFromEmployeePositions 只创建 Position 表记录，不更新 employee.position_id。BuildTree 按 position_id 分组员工，历史员工的 position_id 均为 NULL，导致岗位节点无员工子节点"
  implication: "岗位节点被创建但在树中不可见（无 children 的岗位节点可能被 ECharts 过滤或折叠）。需要 backfill 历史数据的 position_id"
- timestamp: "2026-04-22T18:25:00"
  checked: "department/service.go GetTree() backfill 修复"
  found: "在 GetTree 中获取岗位列表后，新增逻辑：将 position_id 为 NULL 的员工按 position 文本匹配到对应岗位 ID 并更新，然后重新加载员工列表"
  implication: "修复后 BuildTree 的 empByPosID 应有正确数据，岗位节点下会显示员工，岗位节点应在树中可见"

## Resolution
<!-- OVERWRITE as understanding evolves -->

root_cause: |
  根因1（新入职员工岗位不显示）：员工入职流程（CreateEmployee + SubmitRegistration）在创建员工时只设置了 Position 文本字段，从未设置 PositionID 字段，导致 BuildTree 按 position_id 分组时新员工的 PositionID 均为 NULL，无法归入任何岗位节点。

  根因2（编辑框位置错误）：ECharts click 事件提供的 offsetX/offsetY 是相对于 canvas 容器的坐标，但 inline-edit-overlay 使用 position:absolute 相对 .chart-wrapper（无 position:relative）定位，实际定位上下文为 .org-chart-page，其 origin 在 (20, 24) 处。两套坐标系不匹配导致 overlay 错位到页面左上角附近。

  根因3（历史数据岗位不显示）：MigrateFromEmployeePositions 仅根据 Employee.Position 文本创建了 Position 表记录，但从未反向更新 employee.position_id 字段。BuildTree 按 position_id 分组员工时，所有历史员工的 PositionID 仍为 NULL，岗位节点下无员工子节点，因此在树中不可见。

  REGRESSION 根因：vue-echarts 中 offsetX/offsetY 可能始终为 0 或 canvas 左上角坐标，导致 position:fixed 且 left/top 为 0 的编辑框在视口左上角不可见；删除按钮可能因右键事件坐标偏差导致菜单定位错误。

fix: |
  Fix 1 (backend): 在 CreateEmployeeRequest/UpdateEmployeeRequest 中增加 PositionID *int64 字段；
  新增 FindOrCreateByName 方法用于按 Position 文本自动查找/创建岗位并关联 PositionID；
  在 CreateEmployee 和 SubmitRegistration 中调用该方法设置 PositionID；
  修复循环依赖：移除了 position service 对 employee 的依赖，改为由调用方传入员工数据；
  调整 main.go 中各模块的创建顺序（posSvc 移到最前）。

  Fix 2 (frontend): 将 inline-edit-overlay 和 context-menu 的 CSS position 从 absolute 改为 fixed，
  并使用 getBoundingClientRect() 将 ECharts canvas-offset 坐标转为 viewport 坐标。

  Fix 3 (frontend regression fix): 在 bindChartEvents 中将 event.clientX/clientY 作为主坐标源，
  替代可能不可靠的 offsetX/offsetY（fallback 仍保留 offsetX + getBoundingClientRect）。

  Fix 4 (backend data backfill): 在 department/service.go GetTree() 中，获取岗位列表后新增逻辑：
  遍历所有 position_id 为 NULL 的员工，按 position 文本匹配到对应岗位 ID，调用 empRepo.Update 更新 position_id，
  然后重新加载员工列表。确保 MigrateFromEmployeePositions 创建岗位记录后，历史员工也能正确关联岗位。

verification: |
  - go build 编译通过
  - 前端 CSS 修复已应用（position: fixed）
  - event.clientX/clientY 作为主坐标源已应用
  - GetTree 历史数据 backfill 逻辑已应用（position_id 为 NULL 的员工按 position 文本关联到对应岗位）

files_changed:
  - "internal/employee/dto.go": 新增 PositionID 字段到 CreateEmployeeRequest、UpdateEmployeeRequest、EmployeeResponse
  - "internal/employee/service.go": CreateEmployee 增加 PositionID 关联逻辑，UpdateEmployee 增加 PositionID 更新逻辑
  - "internal/employee/registration_service.go": SubmitRegistration 中新员工创建和已存在员工更新均增加 PositionID 关联
  - "internal/position/service.go": 新增 FindOrCreateByName 方法；MigrateFromEmployeePositions 改为接收员工切片（避免循环依赖）；移除 employee 包导入
  - "internal/department/service.go": GetTree 调用处调整为传入员工切片给 MigrateFromEmployeePositions；新增历史数据 backfill 逻辑（position_id 为 NULL 的员工按 position 文本关联到对应岗位）
  - "cmd/server/main.go": 调整模块创建顺序（posSvc 优先），posSvc 只传 Repository 参数，empSvc 和 regSvc 增加 posSvc 参数
  - "frontend/src/views/employee/OrgChart.vue": CSS position 改为 fixed；bindChartEvents 中增加 getChartViewportOffset() 坐标转换；event.clientX/clientY 作为主坐标源