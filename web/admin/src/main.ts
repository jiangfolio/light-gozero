import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { setupPrivilegeDirective } from './directives/privilege'
import router from './router'

import 'antdv-next/dist/reset.css'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
setupPrivilegeDirective(app)
app.mount('#app')
