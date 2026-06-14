import { Link } from 'react-router-dom'
import SimplePageLayout from '@/components/SimplePageLayout'

const COLLECT =
  'Name, e-mail, profile picture, collected by our OAuth provider Kinde. Uploaded images and their metadata (titles, notes, source URL, and so on).'

const USAGE1 = 
  'To provide the service (to authenticate you, to store and display your images) and analytics (check for slow performance).'

const USAGE2 =
  "Identity information (name, e-mail, profile picture) are stored by Bookleaf’s OAuth provider Kinde. All uploaded images and generated thumbnails are stored in Cloudflare R2. Image metadata is stored in Neon’s PostgreSQL."

const USAGE3 =
  "Uploaded images are not shared to any third parties, unless you choose to share them, or you have AI features enabled (a thumbnail version of your image will be sent to Google's Vision API for processing)."

const RETENTION =
  'We keep your identity information as long as your account exists. Uploaded images and their metadata are kept until you delete them or your account.'

const DELETE = 
  'To delete your account, go to Settings → Account → Delete Account. This will delete all your data from our databases and storage, including your identity information, uploaded images, and metadata. This action is irreversible.'

export default function PrivacyPolicyPage() {
  return (
    <SimplePageLayout title="Privacy Policy">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">What we collect</h2>
        <p className="text-muted-foreground">{COLLECT}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">How we use it</h2>
        <p className="text-muted-foreground">{USAGE1}</p>
        <br></br>
        <p className="text-muted-foreground">{USAGE2}</p>
        <br></br>
        <p className="text-muted-foreground">{USAGE3}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Data retention</h2>
        <p className="text-muted-foreground">{RETENTION}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Deleting your data</h2>
        <p className="text-muted-foreground">{DELETE}</p>
      </section>

      <p className="text-muted-foreground">
        See also: <Link to="/ai-notes" className="underline">AI Notes</Link>
      </p>
    </SimplePageLayout>
  )
}
