<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Lock, User } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import dnscLogo from '@/assets/dnsc-logo.svg'
import { login } from '@/stores/auth'

const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const error = ref('')
const submitting = ref(false)

async function handleSubmit() {
  error.value = ''
  submitting.value = true
  try {
    await login(username.value.trim(), password.value)
    const redirect = (route.query.redirect as string) || '/'
    await router.replace(redirect)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Utilizator sau parolă incorectă'
  } finally {
    submitting.value = false
  }
}

function onUserInput(v: string) {
  username.value = v
  error.value = ''
}

function onPassInput(v: string) {
  password.value = v
  error.value = ''
}
</script>

<template>
  <div
    class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/5 via-background to-primary/10 p-4"
  >
    <div class="w-full max-w-sm">
      <div class="flex flex-col items-center mb-8">
        <img :src="dnscLogo" alt="DNSC" class="h-28 w-28 rounded-full mb-2" />
        <p class="text-sm text-muted-foreground">Autentifică-te pentru a continua</p>
      </div>

      <div class="bg-card rounded-2xl shadow-xl border border-border/50 p-8">
        <form class="space-y-5" @submit.prevent="handleSubmit">
          <div class="space-y-2">
            <Label
              html-for="login-username"
              class="text-xs font-medium uppercase tracking-wider text-muted-foreground"
            >
              Utilizator
            </Label>
            <div class="relative">
              <User
                class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/50 pointer-events-none"
              />
              <Input
                id="login-username"
                :model-value="username"
                placeholder="Introdu utilizatorul"
                autocomplete="username"
                class="pl-10 h-11 rounded-xl bg-muted/30 border-border/50 focus-visible:ring-primary/30"
                @update:model-value="onUserInput"
              />
            </div>
          </div>

          <div class="space-y-2">
            <Label
              html-for="login-password"
              class="text-xs font-medium uppercase tracking-wider text-muted-foreground"
            >
              Parolă
            </Label>
            <div class="relative">
              <Lock
                class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/50 pointer-events-none"
              />
              <Input
                id="login-password"
                type="password"
                :model-value="password"
                placeholder="Introdu parola"
                autocomplete="current-password"
                class="pl-10 h-11 rounded-xl bg-muted/30 border-border/50 focus-visible:ring-primary/30"
                @update:model-value="onPassInput"
              />
            </div>
          </div>

          <div
            v-if="error"
            class="bg-destructive/10 text-destructive text-sm px-3 py-2 rounded-lg text-center"
          >
            {{ error }}
          </div>

          <Button
            type="submit"
            class="w-full h-11 rounded-xl font-semibold text-sm shadow-md hover:shadow-lg transition-shadow"
            :disabled="submitting"
          >
            {{ submitting ? 'Se autentifică…' : 'Autentificare' }}
          </Button>
        </form>
      </div>

      <p class="text-center text-xs text-muted-foreground/60 mt-6">
        © 2026 DNSC — Directoratul Național de Securitate Cibernetică
      </p>
    </div>
  </div>
</template>
