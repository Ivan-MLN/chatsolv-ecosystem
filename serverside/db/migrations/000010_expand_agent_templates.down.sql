DELETE FROM agent_templates WHERE id IN ('education_course_cs', 'general_company_cs');
UPDATE agent_templates SET title = 'CS Resto & Kuliner' WHERE id = 'food_beverage_cs';
UPDATE agent_templates SET title = 'CS Layanan & Booking Jasa' WHERE id = 'service_booking_cs';
