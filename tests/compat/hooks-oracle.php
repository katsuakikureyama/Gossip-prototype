<?php
// Owned test adapter. Load the pinned, real WordPress hook engine without a DB.
// This file is executed only inside the network-disabled WordPress container.
error_reporting(E_ALL);
set_error_handler(static function ($severity, $message, $file, $line) {
    throw new ErrorException($message, 0, $severity, $file, $line);
});
define('ABSPATH', '/usr/src/wordpress/');
define('WPINC', 'wp-includes');
require ABSPATH . WPINC . '/version.php';
require ABSPATH . WPINC . '/plugin.php';
$fixture = json_decode(file_get_contents('/fixtures/hooks-v1.json'), true, 512, JSON_THROW_ON_ERROR);
if ($fixture['schema'] !== 'gossip.hook-fixture/v1') {
    throw new RuntimeException('Unsupported fixture schema');
}
$results = [];
foreach ($fixture['cases'] as $case) {
    $hook = 'gossip_fixture_' . $case['id'];
    $trace = [];
    foreach ($case['handlers'] as $index => $handler) {
        add_filter($hook, static function ($value) use ($index, $handler, &$trace) {
            $trace[] = $index;
            return $handler['prefix'] . $value . $handler['suffix'];
        }, $handler['priority'], 1);
    }
    $runs = [];
    for ($run = 0; $run < 2; $run++) {
        $trace = [];
        $value = apply_filters($hook, $case['input']);
        $runs[] = ['value' => $value, 'trace' => $trace];
    }
    remove_all_filters($hook);
    $results[$case['id']] = $runs;
}
echo json_encode([
    'schema' => 'gossip.hook-observation/v1',
    'wordpress' => $wp_version,
    'php' => PHP_VERSION,
    'results' => $results,
], JSON_THROW_ON_ERROR | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
