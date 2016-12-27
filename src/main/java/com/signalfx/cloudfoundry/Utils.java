package com.signalfx.cloudfoundry;

import com.google.common.base.Joiner;

import java.util.List;
import java.util.Map;

/**
 * Various utilities.
 */
class Utils {
    /**
     * Format a map where each key: value pair is on a newline
     * @param map
     * @return formatted string
     */
    static String formatMultilineMap(Map<?, ?> map) {
        return Joiner.on(",\n").withKeyValueSeparator(": ").join(map);
    }

    /**
     * Format a list where each element is on a newline.
     * @param list
     * @return formatted string
     */
    static String formatMultilineList(List<?> list) {
        return "[\n" + Joiner.on("\n").join(list) + "\n]";
    }

    /**
     * Reinterprets a string with contents "null" as null
     * @param s
     * @return
     */
    static String nullishStringToNull(String s) {
        return s != null && s.equals("null") ? null : s;
    }
}
