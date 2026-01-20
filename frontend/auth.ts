import NextAuth from "next-auth"
// import GitHub from "next-auth/providers/github"
import Credentials from "next-auth/providers/credentials"
import { serverApiUrl } from "@/lib/api"
 
export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [
    // GitHub, // Disabled until configured
    Credentials({
      name: "Lingbao Account",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
        captchaId: { label: "Captcha ID", type: "text" },
        captchaCode: { label: "Captcha", type: "text" }
      },
      authorize: async (credentials) => {
        if (!credentials?.username || !credentials?.password) return null;

        try {
          const res = await fetch(serverApiUrl("/api/v1/auth/login"), {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              username: credentials.username,
              password: credentials.password,
              captchaId: credentials.captchaId,
              captchaCode: credentials.captchaCode,
            }),
          });

          if (!res.ok) return null;

          const user = await res.json();
          // User: { token, username, id }
          if (user.token) {
            return {
                id: user.id,
                name: user.username,
                accessToken: user.token,
            }
          }
          return null
        } catch (e) {
          console.error("Auth error:", e)
          return null
        }
      }
    })
  ],
  callbacks: {
    jwt: async ({ token, user }) => {
      if (user) {
        token.accessToken = (user as any).accessToken
        token.id = user.id
      }
      return token
    },
    session: async ({ session, token }) => {
      if (token) {
        (session as any).accessToken = token.accessToken
        // session.user.id = token.id as string
      }
      return session
    },
    authorized: async ({ auth }) => {
      return !!auth
    },
  },
})
