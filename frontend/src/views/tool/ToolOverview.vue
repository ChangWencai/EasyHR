<template>
  <div class="tool-overview">
    <!-- Hero Section -->
    <section class="hero">
      <div class="hero-glow"></div>
      <div class="hero-content">
        <div class="hero-text">
          <h1 class="hero-title">人事工具</h1>
          <p class="hero-subtitle">一站式薪资、社保、个税管理，让繁琐的人事工作变得简单</p>
        </div>
        <div class="hero-stats">
          <div class="hero-stat">
            <div class="hero-stat-icon" style="--stat-color: #A78BFA; --stat-bg: rgba(167,139,250,0.15);">
              <el-icon :size="18"><Coin /></el-icon>
            </div>
            <div class="hero-stat-text">
              <span class="hero-stat-value">--</span>
              <span class="hero-stat-label">本月薪资</span>
            </div>
          </div>
          <div class="hero-stat">
            <div class="hero-stat-icon" style="--stat-color: #60A5FA; --stat-bg: rgba(96,165,250,0.15);">
              <el-icon :size="18"><Umbrella /></el-icon>
            </div>
            <div class="hero-stat-text">
              <span class="hero-stat-value">--</span>
              <span class="hero-stat-label">社保总额</span>
            </div>
          </div>
          <div class="hero-stat">
            <div class="hero-stat-icon" style="--stat-color: #FBBF24; --stat-bg: rgba(251,191,36,0.15);">
              <el-icon :size="18"><Document /></el-icon>
            </div>
            <div class="hero-stat-text">
              <span class="hero-stat-value">--</span>
              <span class="hero-stat-label">待申报</span>
            </div>
          </div>
          <div class="hero-stat">
            <div class="hero-stat-icon" style="--stat-color: #34D399; --stat-bg: rgba(52,211,153,0.15);">
              <el-icon :size="18"><User /></el-icon>
            </div>
            <div class="hero-stat-text">
              <span class="hero-stat-value">--</span>
              <span class="hero-stat-label">在职员工</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Core Tools -->
    <section class="tools-section">
      <div class="section-label">
        <span class="section-dot"></span>
        核心工具
      </div>
      <div class="tools-grid">
        <div
          v-for="(tool, idx) in coreTools"
          :key="tool.path"
          class="tool-card"
          :style="{ '--delay': `${idx * 0.1}s`, '--accent': tool.accent }"
          @click="navigateTo(tool.path)"
        >
          <div class="tool-card-header">
            <div class="tool-icon-wrap" :style="{ background: tool.gradient }">
              <el-icon :size="28" color="#fff"><component :is="tool.icon" /></el-icon>
            </div>
            <div class="tool-badge" :style="{ color: tool.accent, background: tool.badgeBg }">
              {{ tool.badge }}
            </div>
          </div>
          <h3 class="tool-title">{{ tool.title }}</h3>
          <p class="tool-desc">{{ tool.desc }}</p>
          <div class="tool-features">
            <span v-for="f in tool.features" :key="f" class="feature-chip">{{ f }}</span>
          </div>
          <div class="tool-card-action">
            <span>进入工具</span>
            <el-icon :size="14"><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </section>

    <!-- More Tools -->
    <section class="more-section">
      <div class="section-label">
        <span class="section-dot"></span>
        更多工具
      </div>
      <div class="more-grid">
        <div class="more-card glass-card" @click="navigateTo('/tool/email-templates')">
          <div class="more-icon" style="background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%); color: #DC2626;">
            <el-icon :size="20"><Message /></el-icon>
          </div>
          <div class="more-info">
            <span class="more-title">邮箱模板</span>
            <span class="more-desc">管理员工邮件通知模板</span>
          </div>
          <el-icon class="more-arrow"><ArrowRight /></el-icon>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Coin, Umbrella, Document, User, ArrowRight, Message } from '@element-plus/icons-vue'

const router = useRouter()

const coreTools = [
  {
    title: '薪资管理',
    desc: '工资核算、模板配置与工资条发放，一键完成薪资全流程',
    path: '/tool/salary',
    icon: Coin,
    gradient: 'linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%)',
    accent: '#7C3AED',
    badgeBg: 'rgba(124,58,237,0.08)',
    badge: '核心',
    features: ['工资核算', '工资条发送', '薪资模板'],
  },
  {
    title: '社保管理',
    desc: '社保缴纳、参保操作与记录查询，实时掌握社保动态',
    path: '/tool/socialinsurance',
    icon: Umbrella,
    gradient: 'linear-gradient(135deg, #2563EB 0%, #60A5FA 100%)',
    accent: '#2563EB',
    badgeBg: 'rgba(37,99,235,0.08)',
    badge: '重要',
    features: ['政策库', '参保操作', '缴纳记录'],
  },
  {
    title: '个税申报',
    desc: '专项附加扣除、个税计算与申报记录，合规无忧',
    path: '/tool/tax',
    icon: Document,
    gradient: 'linear-gradient(135deg, #D97706 0%, #FBBF24 100%)',
    accent: '#D97706',
    badgeBg: 'rgba(217,119,6,0.08)',
    badge: '必要',
    features: ['专项扣除', '个税计算', '申报记录'],
  },
]

function navigateTo(path: string) {
  router.push(path)
}
</script>

<style scoped lang="scss">
$primary: #7C3AED;
$text-primary: #1A1D2E;
$text-secondary: #5E6278;
$text-muted: #A0A3BD;
$surface: #FFFFFF;
$surface-alt: #F8F9FC;
$border: #E8EBF0;

.tool-overview {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

/* ─── Hero ─── */
.hero {
  position: relative;
  border-radius: 20px;
  overflow: hidden;
  background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 40%, #E0E7FF 100%);
  padding: 40px 36px;
  color: $text-primary;
  border: 1px solid rgba(124,58,237,0.1);
}

.hero-glow {
  position: absolute;
  top: -30%;
  right: -5%;
  width: 260px;
  height: 260px;
  background: radial-gradient(circle, rgba(124,58,237,0.12) 0%, transparent 70%);
  pointer-events: none;
}

.hero::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(124,58,237,0.15), transparent);
}

.hero-content {
  position: relative;
  z-index: 1;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 40px;
}

.hero-title {
  font-size: 28px;
  font-weight: 800;
  margin: 0 0 8px;
  letter-spacing: -0.5px;
  color: #2D1B69;
}

.hero-subtitle {
  font-size: 15px;
  color: $text-secondary;
  margin: 0;
  line-height: 1.5;
}

.hero-stats {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.hero-stat {
  background: rgba(255,255,255,0.7);
  border: 1px solid rgba(255,255,255,0.8);
  border-radius: 14px;
  padding: 14px 18px;
  display: flex;
  align-items: center;
  gap: 12px;
  backdrop-filter: blur(8px);
  min-width: 130px;
  transition: all 0.2s;

  &:hover {
    background: rgba(255,255,255,0.9);
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  }
}

.hero-stat-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: var(--stat-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--stat-color);
  flex-shrink: 0;
}

.hero-stat-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.hero-stat-value {
  font-size: 18px;
  font-weight: 700;
  line-height: 1;
  color: $text-primary;
}

.hero-stat-label {
  font-size: 11px;
  color: $text-muted;
  font-weight: 500;
}

/* ─── Section Label ─── */
.section-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: $text-secondary;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 16px;
}

.section-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: $primary;
}

/* ─── Core Tools ─── */
.tools-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.tool-card {
  background: $surface;
  border: 1px solid $border;
  border-radius: 20px;
  padding: 24px;
  cursor: pointer;
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  gap: 14px;
  animation: cardIn 0.5s cubic-bezier(0.4, 0, 0.2, 1) both;
  animation-delay: var(--delay);

  &:hover {
    border-color: var(--accent);
    box-shadow: 0 8px 32px rgba(0,0,0,0.08), 0 0 0 1px var(--accent);
    transform: translateY(-4px);
  }
}

@keyframes cardIn {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.tool-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tool-icon-wrap {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.tool-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 20px;
  letter-spacing: 0.3px;
}

.tool-title {
  font-size: 17px;
  font-weight: 700;
  color: $text-primary;
  margin: 0;
}

.tool-desc {
  font-size: 13px;
  color: $text-secondary;
  line-height: 1.6;
  margin: 0;
}

.tool-features {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.feature-chip {
  font-size: 11px;
  color: $text-muted;
  background: $surface-alt;
  padding: 3px 10px;
  border-radius: 6px;
  font-weight: 500;
}

.tool-card-action {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  color: var(--accent);
  margin-top: auto;
  padding-top: 8px;
  opacity: 0;
  transform: translateX(-8px);
  transition: all 0.3s ease;

  .tool-card:hover & {
    opacity: 1;
    transform: translateX(0);
  }
}

/* ─── More Tools ─── */
.more-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.more-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.25s ease;
  border: 1px solid $border;

  &:hover {
    border-color: rgba($primary, 0.3);
    box-shadow: 0 4px 16px rgba(0,0,0,0.06);
    transform: translateX(4px);
  }
}

.more-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.more-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.more-title {
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;
}

.more-desc {
  font-size: 12px;
  color: $text-muted;
}

.more-arrow {
  color: $text-muted;
  font-size: 14px;
  transition: all 0.2s;

  .more-card:hover & {
    color: $primary;
    transform: translateX(2px);
  }
}

/* ─── Responsive ─── */
@media (max-width: 1100px) {
  .hero-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 24px;
  }

  .hero-stats {
    flex-wrap: wrap;
  }

  .hero-stat {
    min-width: 120px;
  }

  .tools-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .hero {
    padding: 28px 20px;
  }

  .hero-title {
    font-size: 22px;
  }

  .hero-stats {
    gap: 8px;
  }

  .hero-stat {
    padding: 10px 14px;
    min-width: 0;
    flex: 1;
  }

  .tools-grid {
    grid-template-columns: 1fr;
  }
}
</style>
