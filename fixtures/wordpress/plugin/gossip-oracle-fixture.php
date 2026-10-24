<?php
/**
 * Plugin Name: Gossip Oracle Fixture
 * Description: Checked-in compatibility fixture used only by the local WordPress oracle.
 * Version: 0.1.0
 * License: GPL-2.0-or-later
 */

defined( 'ABSPATH' ) || exit;

add_action(
	'send_headers',
	static function (): void {
		header( 'X-Gossip-Oracle-Hook: send-headers-v1' );
	}
);

add_filter(
	'the_content',
	static function ( string $content ): string {
		return $content . "\n<p class=\"gossip-oracle-plugin-marker\">plugin-content-filter-v1</p>";
	}
);
