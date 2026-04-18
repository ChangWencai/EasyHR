---
phase: "09"
plan: "01"
subsystem: "待办中心-首页环形图与轮播图"
tags: [dashboard, ring-chart, carousel, echarts, home-view]
dependency_graph:
  requires: [dashboard-module, echarts-vue]
  provides: [todo-stats-api, time-limited-stats-api, TodoRingChart, HomeCarousel]
  affects: [HomeView, dashboard-api]
tech_stack:
  added: [vue-echarts-pie, el-carousel]
  patterns: [ring-chart-pattern, carousel-filtering]
key_files:
  created:
    - frontend/src/views/home/components/TodoRingChart.vue
    - frontend/src/views/home/components/HomeCarousel.vue
    - frontend/src/api/carousel.ts
  modified:
    - internal/dashboard/model.go
    - internal/dashboard/repository.go
    - internal/dashboard/service.go
    - internal/dashboard/handler.go
    - internal/dashboard/router.go
    - internal/dashboard/repository_mock.go
    - frontend/src/views/home/HomeView.vue
    - frontend/src/api/dashboard.ts
decisions:
  - TodoItemRecord defined in dashboard package to avoid circular import from internal/todo
  - Table existence check returns (0,0,nil) when todo_items table does not exist yet
  - Percent rounded to 2 decimal places using math.Round
metrics:
  duration: 8min
  completed: "2026-04-19"
---

# Phase 09 Plan 01: 首页环形图 + 轮播图 + 快捷入口扩展 Summary

为管理员首页新增完成率环形图（全部事项/限时任务）、轮播图组件和3个快捷入口（新入职/调薪/考勤），后端提供 todo-stats 和 time-limited-stats 两个 API。

## Completed Tasks

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | 后端环形图统计 API | a6f276f | model.go, repository.go, service.go, handler.go, router.go |
| 2 | 前端环形图组件 + HomeView 更新 | 67ad968 | TodoRingChart.vue, HomeView.vue, dashboard.ts |
| 3 | 轮播图 HomeCarousel 组件 | e44448a | HomeCarousel.vue, carousel.ts, HomeView.vue |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated repository_mock.go for new interface methods**
- **Found during:** Task 1 (build verification)
- **Issue:** MockDashboardRepository and ErrorMockRepository did not implement new GetTodoRingStats/GetTimeLimitedRingStats methods
- **Fix:** Added stub implementations to both mock types
- **Files modified:** internal/dashboard/repository_mock.go
- **Commit:** a6f276f

None otherwise - plan executed as written.

## Key Decisions

1. **TodoItemRecord in dashboard package** - Avoided circular import by defining a mirror struct in dashboard rather than importing from internal/todo
2. **Graceful table-missing handling** - Both ring stats methods return (0, 0, nil) when todo_items table does not exist, matching existing pattern for other optional tables
3. **Percent precision** - Uses math.Round with 2 decimal places for clean display

## Self-Check: PASSED

- All 3 commits verified in git log
- Backend builds successfully (go build ./internal/dashboard/...)
- All grep verifications pass for required symbols
- TodoRingChart.vue, HomeCarousel.vue, carousel.ts created
- All 8 modified files contain expected changes
