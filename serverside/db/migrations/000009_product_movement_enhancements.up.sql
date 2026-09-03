-- 000009_product_movement_enhancements.up.sql
-- ChatSolv Product Movement Database Schema Enhancements

-- 1. Extend business_profiles for structured business intelligence
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS business_type varchar(50) NOT NULL DEFAULT 'products_and_services';
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS target_customer text NOT NULL DEFAULT '';
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS products_services text NOT NULL DEFAULT '';
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS communication_style varchar(50) NOT NULL DEFAULT 'friendly_professional';
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS primary_use_cases jsonb NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS handoff_rules jsonb NOT NULL DEFAULT '{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin", "custom_triggers": []}'::jsonb;
ALTER TABLE business_profiles ADD COLUMN IF NOT EXISTS operating_hours jsonb NOT NULL DEFAULT '{"monday": {"open": "08:00", "close": "17:00", "active": true}, "tuesday": {"open": "08:00", "close": "17:00", "active": true}, "wednesday": {"open": "08:00", "close": "17:00", "active": true}, "thursday": {"open": "08:00", "close": "17:00", "active": true}, "friday": {"open": "08:00", "close": "17:00", "active": true}, "saturday": {"open": "08:00", "close": "15:00", "active": true}, "sunday": {"open": "00:00", "close": "00:00", "active": false}}'::jsonb;

-- 2. Create onboarding_profiles table to persist multi-step onboarding state server-side
CREATE TABLE IF NOT EXISTS onboarding_profiles (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL UNIQUE REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_step integer NOT NULL DEFAULT 1,
    is_completed boolean NOT NULL DEFAULT false,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS onboarding_profiles_workspace_idx ON onboarding_profiles(workspace_id, is_completed);

-- 3. Create workspace_admins table for human CS team management & round-robin rotation
CREATE TABLE IF NOT EXISTS workspace_admins (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    name varchar(120) NOT NULL,
    phone varchar(40) NOT NULL,
    role varchar(40) NOT NULL DEFAULT 'customer_service',
    status varchar(20) NOT NULL DEFAULT 'online' CHECK (status IN ('online', 'busy', 'offline')),
    is_active boolean NOT NULL DEFAULT true,
    rotation_priority integer NOT NULL DEFAULT 0,
    last_assigned_at timestamptz,
    total_handled_today integer NOT NULL DEFAULT 0,
    last_active_date date NOT NULL DEFAULT CURRENT_DATE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, phone)
);
CREATE INDEX IF NOT EXISTS workspace_admins_workspace_idx ON workspace_admins(workspace_id, is_active, status);
CREATE INDEX IF NOT EXISTS workspace_admins_phone_idx ON workspace_admins(phone);

-- 4. Create handoff_requests table for human handoff state tracking & atomic assignment
CREATE TABLE IF NOT EXISTS handoff_requests (
    id uuid PRIMARY KEY,
    short_code varchar(20) NOT NULL UNIQUE,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    customer_phone text NOT NULL,
    reason varchar(80) NOT NULL DEFAULT 'CUSTOMER_REQUEST',
    status varchar(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'accepted', 'resolved', 'expired', 'cancelled')),
    assigned_admin_id uuid REFERENCES workspace_admins(id) ON DELETE SET NULL,
    requested_at timestamptz NOT NULL DEFAULT now(),
    assigned_at timestamptz,
    accepted_at timestamptz,
    resolved_at timestamptz,
    timeout_at timestamptz NOT NULL DEFAULT (now() + interval '2 minutes'),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS handoff_requests_workspace_idx ON handoff_requests(workspace_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS handoff_requests_conversation_idx ON handoff_requests(conversation_id, status);
CREATE INDEX IF NOT EXISTS handoff_requests_short_code_idx ON handoff_requests(short_code);

-- 5. Extend conversations table with waiting_for_admin mode, assigned admin, and handoff references
ALTER TABLE conversations DROP CONSTRAINT IF EXISTS conversations_mode_check;
ALTER TABLE conversations ADD CONSTRAINT conversations_mode_check CHECK (mode IN ('agent', 'waiting_for_admin', 'human'));

ALTER TABLE conversations ADD COLUMN IF NOT EXISTS assigned_admin_id uuid REFERENCES workspace_admins(id) ON DELETE SET NULL;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS current_handoff_id uuid REFERENCES handoff_requests(id) ON DELETE SET NULL;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS handoff_reason varchar(80);
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS human_started_at timestamptz;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS human_ended_at timestamptz;

CREATE INDEX IF NOT EXISTS conversations_assigned_admin_idx ON conversations(workspace_id, assigned_admin_id) WHERE assigned_admin_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS conversations_mode_status_idx ON conversations(workspace_id, mode, status);

-- 6. Create conversation_events table for audit logging and state lifecycle timeline
CREATE TABLE IF NOT EXISTS conversation_events (
    id uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    event_type varchar(80) NOT NULL,
    actor_type varchar(20) NOT NULL CHECK (actor_type IN ('system', 'ai', 'admin', 'customer')),
    actor_id text,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS conversation_events_conv_idx ON conversation_events(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS conversation_events_ws_idx ON conversation_events(workspace_id, created_at DESC);

-- 7. Create agent_templates table for industry-specific AI CS starting templates
CREATE TABLE IF NOT EXISTS agent_templates (
    id varchar(80) PRIMARY KEY,
    industry varchar(80) NOT NULL,
    title varchar(140) NOT NULL,
    description text NOT NULL,
    icon varchar(40) NOT NULL DEFAULT 'Store',
    category varchar(80) NOT NULL,
    default_profile jsonb NOT NULL DEFAULT '{}'::jsonb,
    default_personality jsonb NOT NULL DEFAULT '{}'::jsonb,
    default_use_cases jsonb NOT NULL DEFAULT '[]'::jsonb,
    default_handoff_rules jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_featured boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS agent_templates_industry_idx ON agent_templates(industry, is_featured);

-- Seed initial Agent Templates based on standard industries
INSERT INTO agent_templates (id, industry, title, description, icon, category, default_profile, default_personality, default_use_cases, default_handoff_rules, is_featured)
VALUES
(
    'ecommerce_cs',
    'Toko Online',
    'CS Toko Online Lengkap',
    'Melayani pertanyaan katalog produk, cek ongkir, ketersediaan stok, info pembayaran, status resi, dan komplain pelanggan.',
    'ShoppingBag',
    'E-Commerce & Retail',
    '{"display_name": "CS Toko Online", "description": "Melayani customer toko online dengan ramah, cepat, dan akurat.", "greeting_message": "Halo Kak! Selamat datang di toko kami. Ada produk atau promo yang ingin ditanyakan?", "away_message": "Halo Kak, saat ini toko kami sedang di luar jam operasional. Pesan Kakak sudah kami catat dan akan kami balas segera ya.", "fallback_message": "Mohon tunggu sebentar ya Kak, saya sambungkan ke tim admin kami untuk bantuan lebih lanjut."}',
    '{"bot_name": "CS Online", "role": "Customer Service Toko Online", "tone": "friendly", "communication_style": "conversational", "primary_language": "id", "response_length": "medium", "emoji_usage": "moderate", "greeting_style": "Halo Kak! Selamat datang ya 😊", "closing_style": "Terima kasih banyak Kak!", "custom_instructions": "Bantu pelanggan menemukan produk yang tepat, jelaskan varian warna dan ukuran dengan ramah, berikan info ongkir dan metode pembayaran resmi toko.", "behavior_rules": ["Gunakan panggilan Kak", "Berikan rincian harga dan spesifikasi dengan jelas", "Konfirmasi alamat pengiriman"], "escalation_rules": ["Pelanggan meminta refund", "Bukti transfer bermasalah", "Komplain barang rusak"], "forbidden_topics": ["Membocorkan supplier internal", "Transaksi di luar saluran resmi"], "fallback_behavior": "direct_to_human"}',
    '["Menjawab pertanyaan pelanggan", "Memberikan informasi harga", "Menjelaskan produk", "Membantu memilih produk", "Menjawab pertanyaan stok", "Memberikan informasi pembayaran", "Menjawab pertanyaan pengiriman"]',
    '{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin"}',
    true
),
(
    'food_beverage_cs',
    'Makanan & Minuman',
    'CS Resto & Kuliner',
    'Membantu info menu makanan & minuman, reservasi meja, jam buka resto, lokasi cabang, dan panduan order katering/delivery.',
    'Utensils',
    'Kuliner & Restoran',
    '{"display_name": "CS Restoran", "description": "Asisten reservasi dan informasi menu restoran.", "greeting_message": "Halo Kak! Selamat datang. Mau lihat buku menu spesial kami atau reservasi meja hari ini?", "away_message": "Restoran kami saat ini sedang tutup. Silakan tinggalkan pesan, kami balas saat buka ya Kak.", "fallback_message": "Untuk pemesanan khusus atau konfirmasi meja, saya teruskan ke admin restoran ya Kak."}',
    '{"bot_name": "CS Kuliner", "role": "Restaurant Assistant", "tone": "warm", "communication_style": "conversational", "primary_language": "id", "response_length": "short", "emoji_usage": "moderate", "greeting_style": "Halo Kak! Mau makan apa hari ini? 🍽️", "closing_style": "Selamat menikmati ya Kak!", "custom_instructions": "Jelaskan menu favorit, rekomendasi hidangan, info alergi atau halal, serta panduan reservasi meja.", "behavior_rules": ["Pastikan tanggal dan jumlah orang saat reservasi", "Rekomendasikan menu best-seller"], "escalation_rules": ["Reservasi rombongan besar > 20 orang", "Komplain rasa makanan"], "forbidden_topics": ["Resep rahasia dapur"], "fallback_behavior": "direct_to_human"}',
    '["Menjawab pertanyaan pelanggan", "Memberikan informasi harga", "Menjelaskan produk", "Membantu memilih produk", "Menangani FAQ"]',
    '{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin"}',
    true
),
(
    'beauty_skincare_cs',
    'Beauty / Skincare',
    'Beauty Consultant Assistant',
    'Membantu konsultasi tipe kulit, rekomendasi rangkaian produk skincare, cara pemakaian, dan promo bundling.',
    'Sparkles',
    'Kecantikan & Skincare',
    '{"display_name": "Beauty Assistant", "description": "Konsultasi kecantikan dan rekomendasi skincare terpercaya.", "greeting_message": "Halo Bestie! Mau konsultasi jenis kulit atau cari produk perawatan yang cocok hari ini?", "away_message": "Halo Bestie, admin kami sedang istirahat. Pesan kamu akan segera kami jawab ya!", "fallback_message": "Untuk keluhan kulit yang membutuhkan penanganan khusus, aku sambungkan ke beauty specialist kami ya."}',
    '{"bot_name": "Beauty CS", "role": "Beauty & Skincare Advisor", "tone": "friendly", "communication_style": "conversational", "primary_language": "id", "response_length": "medium", "emoji_usage": "moderate", "greeting_style": "Halo Bestie! ✨", "closing_style": "Semoga kulit makin glowing! 💖", "custom_instructions": "Tanyakan jenis kulit (berminyak, kering, sensitif, berjerawat) sebelum merekomendasikan produk. Jelaskan urutan pemakaian pagi dan malam hari.", "behavior_rules": ["Tanyakan jenis kulit terlebih dahulu", "Berikan instruksi pemakaian yang aman"], "escalation_rules": ["Iritasi parah atau alergi", "Konsultasi medis khusus"], "forbidden_topics": ["Klaim medis berlebihan yang tidak terdaftar BPOM"], "fallback_behavior": "direct_to_human"}',
    '["Menjawab pertanyaan pelanggan", "Menjelaskan produk", "Membantu memilih produk", "Menjawab pertanyaan stok", "Menangani FAQ"]',
    '{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin"}',
    true
),
(
    'service_booking_cs',
    'Jasa',
    'CS Layanan & Booking Jasa',
    'Menjelaskan paket jasa atau servis, estimasi biaya, ketersediaan jadwal teknisi/ahli, dan panduan reservasi layanan.',
    'Briefcase',
    'Layanan & Jasa Profesional',
    '{"display_name": "CS Layanan Jasa", "description": "Membantu jadwal booking dan informasi paket layanan.", "greeting_message": "Halo! Ada yang bisa kami bantu terkait layanan atau reservasi jadwal untuk Anda?", "away_message": "Kantor layanan kami sedang tutup. Silakan infokan kebutuhan Anda, tim kami akan segera menghubungi.", "fallback_message": "Baik, saya hubungkan dengan staf penanggung jawab layanan kami ya."}',
    '{"bot_name": "CS Jasa", "role": "Service Booking Specialist", "tone": "professional", "communication_style": "casual_professional", "primary_language": "id", "response_length": "medium", "emoji_usage": "minimal", "greeting_style": "Halo, selamat datang.", "closing_style": "Terima kasih, senang melayani Anda.", "custom_instructions": "Jelaskan cakupan layanan, estimasi durasi pengerjaan, syarat dan ketentuan, serta catat kebutuhan jadwal pelanggan.", "behavior_rules": ["Tanyakan lokasi pengerjaan jika jasa on-site", "Pastikan tanggal dan waktu yang diinginkan"], "escalation_rules": ["Penawaran harga khusus / enterprise", "Kendala teknisi di lapangan"], "forbidden_topics": ["Menjanjikan garansi di luar ketentuan resmi"], "fallback_behavior": "direct_to_human"}',
    '["Menjawab pertanyaan pelanggan", "Memberikan informasi harga", "Menjelaskan produk", "Menangani FAQ", "Mengarahkan pelanggan ke admin"]',
    '{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin"}',
    true
),
(
    'property_travel_cs',
    'Properti',
    'CS Properti & Konsultasi',
    'Memberikan informasi listing properti, tipe unit, jadwal survey lokasi, dan simulasi skema pembayaran.',
    'Building2',
    'Properti & Real Estate',
    '{"display_name": "CS Properti", "description": "Informasi unit properti dan jadwal kunjungan lokasi.", "greeting_message": "Halo Bapak/Ibu! Selamat datang. Sedang mencari properti hunian atau investasi di area mana?", "away_message": "Pesan Bapak/Ibu telah kami terima. Marketing kami akan segera menghubungi pada jam kerja.", "fallback_message": "Untuk jadwalkan kunjungan lokasi bersama marketing kami, saya hubungkan sekarang ya."}',
    '{"bot_name": "CS Properti", "role": "Property Consultant", "tone": "professional", "communication_style": "formal", "primary_language": "id", "response_length": "detailed", "emoji_usage": "minimal", "greeting_style": "Selamat datang Bapak/Ibu.", "closing_style": "Terima kasih atas kepercayaan Anda.", "custom_instructions": "Berikan spesifikasi unit, keunggulan lokasi, fasilitas sekitar, serta bantu menjadwalkan kunjungan survey lokasi dengan marketing.", "behavior_rules": ["Gunakan sapaan Bapak/Ibu yang sopan", "Catat preferensi tipe unit dan budget"], "escalation_rules": ["Jadwal survey lokasi fix", "Negosiasi harga unit"], "forbidden_topics": ["Menjanjikan persetujuan KPR tanpa verifikasi bank"], "fallback_behavior": "direct_to_human"}',
    '["Menjawab pertanyaan pelanggan", "Memberikan informasi harga", "Menjelaskan produk", "Menangani FAQ", "Mengarahkan pelanggan ke admin"]',
    '{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin"}',
    false
)
ON CONFLICT (id) DO NOTHING;
