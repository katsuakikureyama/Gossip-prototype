#!/bin/sh
set -eu

: "${WORDPRESS_ORACLE_URL:?}"
: "${WORDPRESS_ORACLE_ADMIN_USER:?}"
: "${WORDPRESS_ORACLE_ADMIN_PASSWORD:?}"
: "${WORDPRESS_ORACLE_ADMIN_EMAIL:?}"

if ! wp core is-installed --allow-root 2>/dev/null; then
  wp core install \
    --url="$WORDPRESS_ORACLE_URL" \
    --title="Gossip WordPress Oracle" \
    --admin_user="$WORDPRESS_ORACLE_ADMIN_USER" \
    --admin_password="$WORDPRESS_ORACLE_ADMIN_PASSWORD" \
    --admin_email="$WORDPRESS_ORACLE_ADMIN_EMAIL" \
    --skip-email \
    --allow-root
fi

wp option update permalink_structure '/%postname%/' --allow-root
wp theme activate gossip-oracle --allow-root
wp plugin activate gossip-oracle-fixture --allow-root

if ! wp term get category gossip-oracle --by=slug --allow-root >/dev/null 2>&1; then
  wp term create category 'Gossip Oracle' --slug=gossip-oracle --allow-root
fi

category_id=$(wp term get category gossip-oracle --by=slug --field=term_id --allow-root)

if ! wp post list --post_type=post --name=gossip-oracle-gutenberg --field=ID --allow-root | grep -q .; then
  wp post create /seed/gutenberg-post.html \
    --post_title='Gossip Oracle Gutenberg Fixture' \
    --post_name='gossip-oracle-gutenberg' \
    --post_status=publish \
    --post_category="$category_id" \
    --allow-root
fi

wp option update blogdescription 'Pinned WordPress oracle fixture' --allow-root
echo 'wordpress oracle fixture: ready'
