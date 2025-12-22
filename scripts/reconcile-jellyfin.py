#!/usr/bin/env python3
"""
Script para reconciliar system.xml de Jellyfin con los valores por defecto.
"""
import re
import sys
from pathlib import Path

def get_xml_tag_value(content, tag):
    """Extrae el valor de una etiqueta XML."""
    pattern = f"<{tag}>(.*?)</{tag}>"
    match = re.search(pattern, content, re.DOTALL)
    return match.group(1) if match else ""

def set_xml_tag_value(content, tag, value):
    """Reemplaza el valor de una etiqueta XML."""
    pattern = f"(<{tag}>)(.*?)(</{tag}>)"
    match = re.search(pattern, content, re.DOTALL)
    if not match:
        return content, False
    
    current_value = match.group(2)
    if current_value == value:
        return content, False
    
    new_content = content[:match.start()] + f"<{tag}>{value}</{tag}>" + content[match.end():]
    return new_content, True

def main():
    # Paths
    defaults_path = Path("/home/user/projects/mediacheky/templates/jellyfin.defaults/system.xml")
    target_path = Path("/home/user/projects/mediacheky/volumes/mediacheky-data/services-volumes/jellyfin/system.xml")
    
    if not defaults_path.exists():
        print(f"ERROR: Default system.xml no existe: {defaults_path}")
        sys.exit(1)
    
    if not target_path.exists():
        print(f"ERROR: Target system.xml no existe: {target_path}")
        sys.exit(1)
    
    # Read files
    default_content = defaults_path.read_text()
    target_content = target_path.read_text()
    
    # Tags to reconcile
    tags = [
        "IsStartupWizardCompleted",
        "PreferredMetadataLanguage",
        "MetadataCountryCode",
        "ServerName",
        "UICulture"
    ]
    
    changed = False
    for tag in tags:
        desired_value = get_xml_tag_value(default_content, tag)
        if not desired_value:
            print(f"SKIP: Tag '{tag}' no encontrado en defaults")
            continue
        
        current_value = get_xml_tag_value(target_content, tag)
        print(f"TAG: {tag}")
        print(f"  Current: '{current_value}'")
        print(f"  Desired: '{desired_value}'")
        
        if current_value != desired_value:
            target_content, updated = set_xml_tag_value(target_content, tag, desired_value)
            if updated:
                print(f"  ✓ UPDATED")
                changed = True
            else:
                print(f"  ✗ FAILED TO UPDATE")
        else:
            print(f"  = Already correct")
    
    if changed:
        # Write back
        target_path.write_text(target_content)
        print(f"\n✓ Archivo actualizado: {target_path}")
    else:
        print(f"\n= No se necesitaron cambios")

if __name__ == "__main__":
    main()
