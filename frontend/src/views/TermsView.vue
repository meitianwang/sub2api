<template>
  <div class="flex min-h-screen flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-white">
    <PublicNav />

    <main class="flex-1">
      <article class="terms mx-auto max-w-3xl px-6 py-12 sm:py-16">
        <header class="mb-10 border-b border-gray-200 pb-6 dark:border-gray-800">
          <h1 class="text-3xl font-semibold tracking-tight">服务条款</h1>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-500">
            生效日期：2026 年 4 月 29 日 · 本条款仅以中文版为准
          </p>
        </header>

        <section>
          <p>
            欢迎使用 AIInterface（以下简称"本平台"或"我们"）。本服务条款（以下简称"本条款"）
            是您与 AIInterface 之间就您使用本平台所提供的 API 网关、模型聚合及相关工具
            （以下统称"服务"）所订立的协议。当您注册账号、充值或调用服务时，
            即视为您已阅读并同意本条款的全部内容。如您不同意，请勿使用本服务。
          </p>
        </section>

        <h2>A. 服务说明</h2>
        <p>
          本平台聚合 Claude、GPT、Gemini 等第三方 AI 模型，
          通过统一 API 协议为您提供模型路由、计费、密钥管理等能力。
          本平台不是上游模型的提供方，模型本身的能力、可用性及输出内容由其原厂商决定。
        </p>
        <p>
          您可自行选择是否调用通过本平台暴露的第三方模型功能，本平台不对第三方功能本身负责。
          如您向本平台反馈意见，本平台可自行决定是否采纳，且无须对您承担任何义务。
        </p>

        <h2>B. 账户与注册</h2>
        <p>
          您需通过有效邮箱注册账号。您应妥善保管账号、密码与 API Key，
          不得将凭证转售、外借或以任何形式提供给未授权的第三方。
          因您自身原因导致凭证泄漏所产生的一切费用与后果，由您自行承担。
        </p>
        <p>
          如您发现账号或 API Key 存在被盗用、被攻击或异常调用情况，
          应立即重置密钥并通过下文联系方式通知本平台。
        </p>

        <h2>C. 充值与计费</h2>
        <p>
          本平台采用预付费 + 按 token 计费模式。您需先向账号充值余额，
          调用服务时按所用模型的实时价格扣减余额。各模型的输入 / 输出单价
          以"模型与定价"页面公示为准，本平台保留根据上游成本变动调整定价的权利，
          调整将提前在站内公告。
        </p>
        <p>
          您应及时支付您账号下产生的所有费用。余额耗尽时，
          本平台有权暂停您对服务的访问，直至您完成续费。
          您对本平台支付的款项不包含因您使用服务可能产生的任何税费、关税或类似费用，
          相关税费由您自行承担。
        </p>

        <h2>D. 退款政策</h2>
        <ul>
          <li>
            <strong>退款范围</strong>：因您自身原因（包括但不限于不再需要服务、
            使用计划调整、对模型效果不满意等）申请退款的，本平台有权拒绝。
            因本平台原因导致服务长时间不可用，且不可用期间您账户持续存在余额的，
            可申请按比例退款。
          </li>
          <li>
            <strong>退款方式</strong>：充值后 90 天内的退款，将原路退回至原支付渠道；
            超过 90 天的，因支付渠道无法原路退回，您需另行提供与付款人身份一致的收款方式。
          </li>
          <li>
            <strong>审核与材料</strong>：所有退款申请均需经本平台审核。您需提交：
            (a) 原始支付凭证（支付截图、订单号等）；
            (b) 与付款账户一致的身份信息；
            (c) 无法原路退回时，与付款人身份一致的收款账户信息。
            本平台有权要求补充额外材料。
          </li>
          <li>
            <strong>反欺诈</strong>：为防止欺诈、洗钱及滥用补贴等行为，
            本平台保留对退款申请进行额外审查的权利，
            包括但不限于核验付款人与收款人的身份一致性。
            如发现申请存在欺诈嫌疑，本平台有权拒绝退款并暂停相关账号。
          </li>
          <li>
            <strong>最终解释权</strong>：本平台保留对退款政策的最终解释权，
            并有权拒绝不符合退款条件的申请。
          </li>
        </ul>

        <h2>E. 使用规范</h2>
        <p>您不得（且不得协助第三方）通过本平台从事下列行为：</p>
        <ul>
          <li>违反中华人民共和国法律法规或公序良俗；</li>
          <li>生成、传播涉及暴力、色情、恐怖、歧视、欺诈、虚假信息等违法或有害内容；</li>
          <li>逆向工程、抓取或复制本平台的服务实现；</li>
          <li>构建与本平台直接竞争的同类聚合服务，或转售本平台的接口；</li>
          <li>对本平台或上游服务发起拒绝服务攻击、绕过限流或其他恶意调用；</li>
          <li>侵犯他人知识产权、隐私权、名誉权或其他合法权益；</li>
          <li>其他可能损害本平台、其他用户或上游模型提供方利益的行为。</li>
        </ul>
        <p>
          您应自行评估模型输出是否适合您的用例，并在必要时进行人工复核。
          模型输出可能包含不准确、过时或具有偏见的信息，您不应在未经核实的情况下
          将其作为事实依据，亦不应将其用于医疗、法律、金融等可能影响人身或财产安全的关键决策。
        </p>

        <h2>F. 知识产权</h2>
        <p>
          在适用法律允许的范围内：
          您保留对您提交至服务的输入（"输入"）的全部权利；
          您拥有服务针对您输入生成的输出（"输出"）。
          本平台不会使用您的输入或输出训练任何模型。
          上述权利分配以您不违反本条款为前提。
        </p>
        <p>
          本平台的代码、商标、品牌标识、文档及网站设计等知识产权归本平台所有，
          除本条款明确授权外，您不得复制、修改或用于其他用途。
        </p>

        <h2>G. 隐私</h2>
        <p>
          本平台仅在为您提供服务所必需的范围内收集和处理您的账户信息、调用日志与计费数据。
          您的请求与响应内容不会被用于模型训练，也不会向无关第三方披露，
          但本平台可能将必要信息（如请求体、模型 ID）转发给上游模型提供方以完成调用，
          此类转发受上游服务自身的隐私政策约束。
        </p>

        <h2>H. 服务可用性与免责</h2>
        <p>
          本平台以"现状"和"可用"方式提供服务，不就以下事项作出任何明示或暗示的保证：
          (a) 服务或输出绝对准确、完整、无错误或不会中断；
          (b) 任何第三方模型或接口的可用性、质量或合规性；
          (c) 服务能够满足您的特定业务需求。
        </p>
        <p>
          因下列情形导致的服务中断或异常，本平台不承担责任：
          (a) 不可抗力（如自然灾害、战争、政府行为）；
          (b) 上游模型提供方限流、封禁、调价或停止服务；
          (c) 您使用的网络运营商、设备或浏览器问题；
          (d) 您违反本条款导致的暂停。
        </p>

        <h2>I. 责任限制</h2>
        <p>
          在适用法律允许的最大范围内，本平台及其关联方对因本条款或您使用服务而产生的任何
          间接的、附带的、特殊的、惩罚性的损失（包括但不限于利润损失、业务损失、
          数据丢失或商誉损失）不承担责任。
          无论基于何种原因，本平台对您的累计赔偿责任总额不超过您在最近 12 个月内为服务
          实际支付的费用金额。
        </p>

        <h2>J. 终止与暂停</h2>
        <p>
          您可随时停止使用服务并注销账号；账号注销后剩余余额按"退款政策"处理。
          如您严重违反本条款（包括但不限于"使用规范"中列出的禁止行为），
          本平台有权立即暂停或终止您对服务的访问，并保留追究您法律责任的权利。
        </p>
        <p>
          在以下情形下，本平台亦可暂停您对服务的访问：
          (a) 本平台合理判断您的使用对服务或其他用户构成风险；
          (b) 上游服务因任何原因停止向本平台提供能力；
          (c) 本平台向您提供服务被适用法律禁止。
        </p>

        <h2>K. 条款修改</h2>
        <p>
          本平台可不时更新本条款。修改后的条款将在本页面发布，
          重大修改将提前 30 天在站内公告或通过您预留的邮箱通知您。
          修改后的条款发布生效之日起，您继续使用服务即视为接受修改后的条款。
          如您不接受修改，应停止使用并申请注销账号。
        </p>

        <h2>L. 适用法律与争议解决</h2>
        <p>
          本条款的订立、效力、解释、履行及争议解决均适用中华人民共和国法律
          （不含港澳台地区法律）。如您与本平台就本条款产生争议，
          双方应首先友好协商解决；协商不成的，任何一方可向 [城市] 有管辖权的人民法院提起诉讼。
        </p>

        <h2>M. 联系方式</h2>
        <p>
          如您对本条款有任何疑问，或需提交退款、违规举报、隐私相关请求，请通过以下方式联系我们：
        </p>
        <ul>
          <li>邮箱：<code class="font-mono text-sm">support@aiinterface.store</code></li>
        </ul>
      </article>
    </main>

    <footer class="border-t border-gray-200 dark:border-gray-800">
      <div class="mx-auto flex max-w-5xl flex-col gap-3 px-6 py-8 text-xs text-gray-500 dark:text-gray-500 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} AIInterface. 保留所有权利。</p>
        <div class="flex items-center gap-5">
          <router-link to="/home" class="hover:text-gray-900 dark:hover:text-white">首页</router-link>
          <router-link to="/docs" class="hover:text-gray-900 dark:hover:text-white">文档</router-link>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import PublicNav from '@/components/common/PublicNav.vue'

const currentYear = computed(() => new Date().getFullYear())
</script>

<style scoped>
.terms { @apply text-gray-800 dark:text-gray-200 leading-relaxed; }
.terms h2 {
  @apply text-xl font-semibold text-gray-900 dark:text-white mt-10 mb-4 pb-2;
  @apply border-b border-gray-100 dark:border-gray-900;
}
.terms p { @apply mb-4 leading-7; }
.terms ul { @apply mb-4 list-disc pl-6 space-y-2; }
.terms li { @apply leading-7; }
.terms strong { @apply font-semibold text-gray-900 dark:text-white; }
.terms code { @apply bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-1.5 py-0.5 rounded text-sm; }
</style>
